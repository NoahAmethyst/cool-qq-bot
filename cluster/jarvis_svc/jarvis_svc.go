// Package jarvis_svc 封装 Jarvis gRPC 客户端及其生命周期。
package jarvis_svc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/NoahAmethyst/go-cqhttp/modules/config"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/jarvis_pb"
)

var (
	mu sync.RWMutex

	connection       *grpc.ClientConn
	client           jarvis_pb.JarvisServiceClient
	chatTimeout      time.Duration
	ingestTimeout    time.Duration
	summaryTimeout   time.Duration
	wallStreetUserID string
)

const qqOutputPolicy = `QQ输出约束：
返回适合直接发送到QQ的普通文本。
可以使用emoji和普通中文标点、普通符号突出标题和重点。
不允许使用Markdown格式，不要使用Markdown标题、项目符号、加粗、斜体、删除线、代码块、表格、引用或Markdown链接。
可以用简短自然段组织内容。`

// Init 建立并保存一个可复用的 Jarvis gRPC 连接。
func Init(cfg config.Jarvis) error {
	target := strings.TrimSpace(cfg.GRPCTarget)
	userID := strings.TrimSpace(cfg.WallStreetUserID)
	if target == "" {
		return fmt.Errorf("jarvis grpc target is empty")
	}
	if userID == "" {
		return fmt.Errorf("jarvis wall street user id is empty")
	}
	if cfg.ConnectTimeoutSeconds <= 0 {
		return fmt.Errorf("jarvis connect timeout must be positive")
	}
	if cfg.ChatTimeoutSeconds <= 0 {
		return fmt.Errorf("jarvis chat timeout must be positive")
	}
	if cfg.IngestTimeoutSeconds <= 0 {
		return fmt.Errorf("jarvis ingest timeout must be positive")
	}
	if cfg.SummaryTimeoutSeconds <= 0 {
		return fmt.Errorf("jarvis summary timeout must be positive")
	}

	mu.Lock()
	defer mu.Unlock()
	if connection != nil {
		return fmt.Errorf("jarvis grpc client is already initialized")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.ConnectTimeoutSeconds)*time.Second,
	)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return err
	}

	connection = conn
	client = jarvis_pb.NewJarvisServiceClient(conn)
	chatTimeout = time.Duration(cfg.ChatTimeoutSeconds) * time.Second
	ingestTimeout = time.Duration(cfg.IngestTimeoutSeconds) * time.Second
	summaryTimeout = time.Duration(cfg.SummaryTimeoutSeconds) * time.Second
	wallStreetUserID = userID
	log.Info("Jarvis gRPC 连接成功")
	return nil
}

type clientSnapshot struct {
	client           jarvis_pb.JarvisServiceClient
	chatTimeout      time.Duration
	ingestTimeout    time.Duration
	summaryTimeout   time.Duration
	wallStreetUserID string
}

func snapshot() (clientSnapshot, error) {
	mu.RLock()
	defer mu.RUnlock()
	if client == nil {
		return clientSnapshot{}, fmt.Errorf("jarvis grpc client is not initialized")
	}
	return clientSnapshot{
		client:           client,
		chatTimeout:      chatTimeout,
		ingestTimeout:    ingestTimeout,
		summaryTimeout:   summaryTimeout,
		wallStreetUserID: wallStreetUserID,
	}, nil
}

func withQQOutputPolicy(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return qqOutputPolicy
	}
	return message + "\n\n" + qqOutputPolicy
}

func chatWithTimeout(s clientSnapshot, timeout time.Duration, message, userID string) (*jarvis_pb.ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.client.Chat(ctx, &jarvis_pb.ChatRequest{
		Message: withQQOutputPolicy(message),
		UserId:  userID,
	})
}

func generateWithTimeout(s clientSnapshot, timeout time.Duration, prompt, userID, operation string) (*jarvis_pb.GenerateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.client.Generate(ctx, &jarvis_pb.GenerateRequest{
		Prompt:    withQQOutputPolicy(prompt),
		UserId:    userID,
		Operation: operation,
	})
}

// Chat 向 Jarvis 发起普通对话，由 Jarvis 管理模型和对话生命周期。
func Chat(message, userID string) (*jarvis_pb.ChatResponse, error) {
	s, err := snapshot()
	if err != nil {
		return nil, err
	}
	return chatWithTimeout(s, s.chatTimeout, message, userID)
}

// Summary 使用无记忆生成接口请求 Jarvis 生成华尔街见闻总结。
func Summary(prompt string) (*jarvis_pb.GenerateResponse, error) {
	s, err := snapshot()
	if err != nil {
		return nil, err
	}
	return generateWithTimeout(s, s.summaryTimeout, prompt, s.wallStreetUserID, "wallstreet_summary_v2")
}

// Weather 使用无记忆生成接口请求 Jarvis 生成定时天气报告。
func Weather(prompt, userID string) (*jarvis_pb.GenerateResponse, error) {
	s, err := snapshot()
	if err != nil {
		return nil, err
	}
	return generateWithTimeout(s, s.chatTimeout, prompt, userID, "weather_lookup")
}

// Ingest 将华尔街见闻材料写入 Jarvis 知识库。
func Ingest(content, sourceURL string) (*jarvis_pb.IngestResponse, error) {
	s, err := snapshot()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.ingestTimeout)
	defer cancel()
	return s.client.Ingest(ctx, &jarvis_pb.IngestRequest{
		Content:   content,
		SourceUrl: sourceURL,
		UserId:    s.wallStreetUserID,
	})
}

// WallStreetUserID 返回华尔街见闻知识库使用的稳定身份。
func WallStreetUserID() string {
	mu.RLock()
	defer mu.RUnlock()
	return wallStreetUserID
}

// Close 关闭 Jarvis gRPC 连接。
func Close() error {
	mu.Lock()
	defer mu.Unlock()
	if connection == nil {
		return nil
	}
	conn := connection
	connection = nil
	client = nil
	chatTimeout = 0
	ingestTimeout = 0
	summaryTimeout = 0
	wallStreetUserID = ""
	return conn.Close()
}
