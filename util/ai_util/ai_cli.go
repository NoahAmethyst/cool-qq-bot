package ai_util

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/NoahAmethyst/go-cqhttp/constant"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
)

var (
	openaiMu   sync.RWMutex
	openaiCli  *openai.Client
	openaiKey  string
	openaiOnce sync.Once
)

func initCli() {
	openaiOnce.Do(func() {
		openaiMu.Lock()
		defer openaiMu.Unlock()
		openaiKey = strings.TrimSpace(os.Getenv(constant.OPENAI_API_KEY))
		openaiCli = newOpenAIClient(openaiKey)
	})
}

func newOpenAIClient(key string) *openai.Client {
	if key == "" {
		return nil
	}
	log.Info("init OpenAI image client")
	if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
		return openai.NewClient(key)
	}
	if proxy := strings.TrimSpace(os.Getenv(constant.OPENAI_PROXY)); proxy != "" {
		config := openai.DefaultConfig(key)
		config.BaseURL = fmt.Sprintf("https://%s/v1", proxy)
		return openai.NewClientWithConfig(config)
	}
	return openai.NewClient(key)
}

func SetOpenaiKey(key string) {
	initCli()
	openaiMu.Lock()
	defer openaiMu.Unlock()
	openaiKey = strings.TrimSpace(key)
	openaiCli = newOpenAIClient(openaiKey)
}

func openAIClient() *openai.Client {
	initCli()
	openaiMu.RLock()
	defer openaiMu.RUnlock()
	return openaiCli
}
