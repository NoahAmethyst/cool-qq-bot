package coolq

import (
	"context"
	"math"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/cluster/rpc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type recordingGoldAlertSender struct {
	groupIDs        []int64
	privateIDs      []int64
	groupContents   []string
	privateContents []string
}

func (s *recordingGoldAlertSender) SendGroupMessage(id int64, msg *message.SendingMessage) int32 {
	s.groupIDs = append(s.groupIDs, id)
	s.groupContents = append(s.groupContents, goldAlertText(msg))
	return 1
}

func (s *recordingGoldAlertSender) SendPrivateMessage(id int64, _ int64, msg *message.SendingMessage) int32 {
	s.privateIDs = append(s.privateIDs, id)
	s.privateContents = append(s.privateContents, goldAlertText(msg))
	return 1
}

func goldAlertText(msg *message.SendingMessage) string {
	if msg == nil || len(msg.Elements) != 1 {
		return ""
	}
	text, _ := msg.Elements[0].(*message.TextElement)
	if text == nil {
		return ""
	}
	return text.Content
}

type goldPriceSpiderServer struct {
	spider_pb.UnimplementedSpiderServiceServer

	mu    sync.Mutex
	calls int
	price float32
	err   error

	blockStarted chan struct{}
	blockRelease chan struct{}
	blockOnce    sync.Once
}

func (s *goldPriceSpiderServer) Finance(context.Context, *spider_pb.SpiderReq) (*spider_pb.SpiderResp, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	if s.err != nil {
		return nil, s.err
	}
	if s.blockStarted != nil {
		s.blockOnce.Do(func() {
			close(s.blockStarted)
		})
		<-s.blockRelease
	}
	return &spider_pb.SpiderResp{
		FinanceResp: &spider_pb.FinanceResp{FloatValue: s.price},
	}, nil
}

func (s *goldPriceSpiderServer) callCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func startGoldPriceSpiderServer(t *testing.T, price float32) *goldPriceSpiderServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for fake spider service: %v", err)
	}

	server := grpc.NewServer()
	fake := &goldPriceSpiderServer{price: price}
	spider_pb.RegisterSpiderServiceServer(server, fake)
	go func() {
		_ = server.Serve(listener)
	}()

	t.Setenv("SPIDER_SERVICE_ADDR", listener.Addr().String())
	rpc.InitGrpcCli(rpc.CliSpider)
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
		if conn := rpc.GetConn(rpc.CliSpider); conn != nil {
			_ = conn.Close()
		}
	})
	return fake
}

func TestSendGoldPriceAlertFansOutToGroupAndPrivateSubscribers(t *testing.T) {
	sender := &recordingGoldAlertSender{}

	sendGoldPriceAlert(sender, []int64{101, 102}, []int64{201}, "gold alert")

	if got, want := sender.groupIDs, []int64{101, 102}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("group recipients = %v, want %v", got, want)
	}
	if got, want := sender.privateIDs, []int64{201}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("private recipients = %v, want %v", got, want)
	}
	for _, got := range append(sender.groupContents, sender.privateContents...) {
		if got != "gold alert" {
			t.Fatalf("alert content = %q, want %q", got, "gold alert")
		}
	}
}

func TestGoldPriceReportAlertsGroupAndPrivateSubscribersFromOneObservation(t *testing.T) {
	fake := startGoldPriceSpiderServer(t, 110)
	sender := &recordingGoldAlertSender{}
	bot := &CQBot{
		state: &State{
			globalState: &globalState{
				Data: map[string]string{"last_gold_price": "100"},
			},
		},
	}

	bot.reportGoldPriceFluctuation([]int64{101, 102}, []int64{201}, sender)

	if got := fake.callCount(); got != 1 {
		t.Fatalf("gold price fetch count = %d, want 1", got)
	}
	if got, want := sender.groupIDs, []int64{101, 102}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("group recipients = %v, want %v", got, want)
	}
	if got, want := sender.privateIDs, []int64{201}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("private recipients = %v, want %v", got, want)
	}
	if sender.groupContents[0] != sender.privateContents[0] {
		t.Fatalf("group alert %q differs from private alert %q", sender.groupContents[0], sender.privateContents[0])
	}
	if got, _ := bot.state.globalState.GetData("last_gold_price"); got != "110" {
		t.Fatalf("recorded gold price = %q, want %q", got, "110")
	}
}

func TestGoldFluctuationReporterFetchesAndRecordsOncePerRun(t *testing.T) {
	fake := startGoldPriceSpiderServer(t, 105)
	bot := &CQBot{
		state: &State{
			reportState: &reportState{
				groups:   map[int64]struct{}{},
				privates: map[int64]struct{}{},
			},
			globalState: &globalState{
				Data: map[string]string{"last_gold_price": "100"},
			},
		},
	}

	bot.GoldFluctuationReporter("unused").Report()

	if got := fake.callCount(); got != 1 {
		t.Errorf("gold price fetch count = %d, want 1 per scheduled run", got)
	}
	if got, ok := bot.state.globalState.GetData("last_gold_price"); !ok || got != "105" {
		t.Errorf("recorded gold price = %q, present = %v; want %q, true", got, ok, "105")
	}
}

func TestGoldFluctuationReporterSkipsOverlappingRun(t *testing.T) {
	fake := startGoldPriceSpiderServer(t, 105)
	fake.blockStarted = make(chan struct{})
	fake.blockRelease = make(chan struct{})
	bot := &CQBot{
		state: &State{
			reportState: &reportState{
				groups:   map[int64]struct{}{},
				privates: map[int64]struct{}{},
			},
			globalState: &globalState{
				Data: map[string]string{"last_gold_price": "100"},
			},
		},
	}
	job := bot.GoldFluctuationReporter("unused")

	firstDone := make(chan struct{})
	go func() {
		job.Report()
		close(firstDone)
	}()
	select {
	case <-fake.blockStarted:
	case <-time.After(time.Second):
		t.Fatal("first gold report did not reach finance service")
	}

	secondDone := make(chan struct{})
	go func() {
		job.Report()
		close(secondDone)
	}()
	select {
	case <-secondDone:
	case <-time.After(200 * time.Millisecond):
		close(fake.blockRelease)
		<-firstDone
		<-secondDone
		t.Fatal("overlapping gold report did not return promptly")
	}

	if got := fake.callCount(); got != 1 {
		close(fake.blockRelease)
		<-firstDone
		t.Fatalf("gold price fetch count during overlap = %d, want 1", got)
	}
	close(fake.blockRelease)
	<-firstDone
}

func TestGoldFluctuationReporterKeepsBaselineOnRPCError(t *testing.T) {
	fake := startGoldPriceSpiderServer(t, 0)
	fake.err = status.Error(codes.Unavailable, "spider unavailable")
	bot := &CQBot{
		state: &State{
			reportState: &reportState{
				groups:   map[int64]struct{}{},
				privates: map[int64]struct{}{},
			},
			globalState: &globalState{
				Data: map[string]string{"last_gold_price": "100"},
			},
		},
	}

	bot.GoldFluctuationReporter("unused").Report()

	if got, _ := bot.state.globalState.GetData("last_gold_price"); got != "100" {
		t.Fatalf("recorded gold price = %q after RPC error, want unchanged %q", got, "100")
	}
}

func TestGoldFluctuationReporterRejectsNonFinitePrices(t *testing.T) {
	tests := []struct {
		name  string
		price float32
	}{
		{name: "NaN", price: float32(math.NaN())},
		{name: "positive infinity", price: float32(math.Inf(1))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startGoldPriceSpiderServer(t, tt.price)
			bot := &CQBot{
				state: &State{
					reportState: &reportState{
						groups:   map[int64]struct{}{},
						privates: map[int64]struct{}{},
					},
					globalState: &globalState{
						Data: map[string]string{"last_gold_price": "100"},
					},
				},
			}

			bot.GoldFluctuationReporter("unused").Report()

			if got, _ := bot.state.globalState.GetData("last_gold_price"); got != "100" {
				t.Fatalf("recorded gold price = %q after %s, want unchanged %q", got, tt.name, "100")
			}
		})
	}
}

func TestGoldPriceFluctuationThreshold(t *testing.T) {
	float32LastPrice := float64(float32(100.01))
	float32LatestPrice := float64(float32(float32(100.01) * 0.9))
	tests := []struct {
		name        string
		lastPrice   float64
		latestPrice float64
		wantPercent float64
		wantAlert   bool
	}{
		{name: "rise below threshold", lastPrice: 100, latestPrice: 109.99, wantPercent: 9.99},
		{name: "rise reaches threshold", lastPrice: 100, latestPrice: 110, wantPercent: 10, wantAlert: true},
		{name: "fall below threshold", lastPrice: 100, latestPrice: 90.01, wantPercent: -9.99},
		{name: "fall reaches threshold", lastPrice: 100, latestPrice: 90, wantPercent: -10, wantAlert: true},
		{
			name:        "float32 fall reaches threshold",
			lastPrice:   float32LastPrice,
			latestPrice: float32LatestPrice,
			wantPercent: -10,
			wantAlert:   true,
		},
		{name: "invalid previous price", lastPrice: 0, latestPrice: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPercent, gotAlert := goldPriceFluctuation(tt.lastPrice, tt.latestPrice)
			if diff := gotPercent - tt.wantPercent; diff < -0.0001 || diff > 0.0001 {
				t.Errorf("fluctuation percentage = %.4f, want %.4f", gotPercent, tt.wantPercent)
			}
			if gotAlert != tt.wantAlert {
				t.Errorf("alert = %v, want %v", gotAlert, tt.wantAlert)
			}
		})
	}
}
