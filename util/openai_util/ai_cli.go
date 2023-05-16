package openai_util

import (
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/sashabaranov/go-openai"
	"os"
	"time"
)

var openaiCli *openai.Client

func initCli() {
	apiKey := os.Getenv(constant.OPENAI_API_KEY)
	if openaiCli == nil {
		if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
			openaiCli = openai.NewClient(apiKey)
		} else {
			config := openai.DefaultConfig(apiKey)
			config.HTTPClient.Timeout = time.Minute * 60
			config.BaseURL = "https://open.aiproxy.xyz/v1"
			openaiCli = openai.NewClientWithConfig(config)
		}
	}
}
