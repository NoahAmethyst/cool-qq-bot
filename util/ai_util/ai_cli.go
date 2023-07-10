package ai_util

import (
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/sashabaranov/go-openai"
	"os"
	"time"
)

var openaiCli *openai.Client

var openaiKey string

func initCli() {
	if len(openaiKey) == 0 {
		openaiKey = os.Getenv(constant.OPENAI_API_KEY)
	}

	if openaiCli == nil {
		if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
			openaiCli = openai.NewClient(openaiKey)
		} else {
			config := openai.DefaultConfig(openaiKey)
			config.HTTPClient.Timeout = time.Minute * 120
			config.BaseURL = "https://open.aiproxy.xyz/v1"
			openaiCli = openai.NewClientWithConfig(config)
		}
	}
}

func SetOpenaiKey(key string) {
	openaiKey = key
}
