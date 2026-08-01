package jarvis_svc

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/jarvis_pb"
	"google.golang.org/grpc"
)

type fakeJarvisClient struct {
	jarvis_pb.UnimplementedJarvisServiceServer
	generateReq *jarvis_pb.GenerateRequest
}

func (f *fakeJarvisClient) Chat(context.Context, *jarvis_pb.ChatRequest, ...grpc.CallOption) (*jarvis_pb.ChatResponse, error) {
	return &jarvis_pb.ChatResponse{Answer: "chat"}, nil
}

func (f *fakeJarvisClient) Generate(_ context.Context, req *jarvis_pb.GenerateRequest, _ ...grpc.CallOption) (*jarvis_pb.GenerateResponse, error) {
	f.generateReq = req
	return &jarvis_pb.GenerateResponse{Text: "generated"}, nil
}

func (f *fakeJarvisClient) Ingest(context.Context, *jarvis_pb.IngestRequest, ...grpc.CallOption) (*jarvis_pb.IngestResponse, error) {
	return &jarvis_pb.IngestResponse{Success: true}, nil
}

func (f *fakeJarvisClient) GetMemory(context.Context, *jarvis_pb.MemoryRequest, ...grpc.CallOption) (*jarvis_pb.MemoryResponse, error) {
	return &jarvis_pb.MemoryResponse{}, nil
}

func (f *fakeJarvisClient) DeleteMemory(context.Context, *jarvis_pb.MemoryRequest, ...grpc.CallOption) (*jarvis_pb.DeleteResponse, error) {
	return &jarvis_pb.DeleteResponse{Success: true}, nil
}

func installFakeJarvisClient(t *testing.T) *fakeJarvisClient {
	t.Helper()
	mu.Lock()
	defer mu.Unlock()

	oldClient := client
	oldChatTimeout := chatTimeout
	oldIngestTimeout := ingestTimeout
	oldSummaryTimeout := summaryTimeout
	oldWallStreetUserID := wallStreetUserID

	fake := &fakeJarvisClient{}
	client = fake
	chatTimeout = time.Second
	ingestTimeout = time.Second
	summaryTimeout = time.Second
	wallStreetUserID = "go-cqhttp:wallstreet"

	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()
		client = oldClient
		chatTimeout = oldChatTimeout
		ingestTimeout = oldIngestTimeout
		summaryTimeout = oldSummaryTimeout
		wallStreetUserID = oldWallStreetUserID
	})
	return fake
}

func TestSummaryUsesGenerateWithoutChatMemory(t *testing.T) {
	fake := installFakeJarvisClient(t)

	response, err := Summary("summary prompt")
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if response.GetText() != "generated" {
		t.Fatalf("Summary() text = %q, want generated", response.GetText())
	}
	if fake.generateReq == nil {
		t.Fatal("Summary() did not call Generate")
	}
	if fake.generateReq.GetUserId() != "go-cqhttp:wallstreet" {
		t.Fatalf("Generate user_id = %q", fake.generateReq.GetUserId())
	}
	if fake.generateReq.GetOperation() != "wallstreet_summary_v2" {
		t.Fatalf("Generate operation = %q", fake.generateReq.GetOperation())
	}
	if fake.generateReq.GetDisableTools() {
		t.Fatal("Generate disable_tools = true, want default false")
	}
	if fake.generateReq.GetDisableRag() {
		t.Fatal("Generate disable_rag = true, want default false")
	}
	if !strings.Contains(fake.generateReq.GetPrompt(), "summary prompt") {
		t.Fatalf("Generate prompt = %q, want original prompt included", fake.generateReq.GetPrompt())
	}
	if !strings.Contains(fake.generateReq.GetPrompt(), "QQ输出约束") {
		t.Fatalf("Generate prompt = %q, want QQ output policy included", fake.generateReq.GetPrompt())
	}
}

func TestWeatherUsesGenerateWithoutChatMemory(t *testing.T) {
	fake := installFakeJarvisClient(t)

	response, err := Weather("weather prompt", "go-cqhttp:weather:test")
	if err != nil {
		t.Fatalf("Weather() error = %v", err)
	}
	if response.GetText() != "generated" {
		t.Fatalf("Weather() text = %q, want generated", response.GetText())
	}
	if fake.generateReq == nil {
		t.Fatal("Weather() did not call Generate")
	}
	if fake.generateReq.GetUserId() != "go-cqhttp:weather:test" {
		t.Fatalf("Generate user_id = %q", fake.generateReq.GetUserId())
	}
	if fake.generateReq.GetOperation() != "weather_lookup" {
		t.Fatalf("Generate operation = %q", fake.generateReq.GetOperation())
	}
	if fake.generateReq.GetDisableTools() {
		t.Fatal("Generate disable_tools = true, want default false")
	}
	if fake.generateReq.GetDisableRag() {
		t.Fatal("Generate disable_rag = true, want default false")
	}
}
