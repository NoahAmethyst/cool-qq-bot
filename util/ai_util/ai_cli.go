package ai_util

import (
	"context"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/sashabaranov/go-openai"
	"os"
	"sync"
	"time"
)

var openaiCli *openai.Client
var chimeraCli *openai.Client

var openaiKey string
var chimeraKey string

var changeSignal = make(chan struct{}, 1)
var once sync.Once

func OpenAiModels() (openai.ModelsList, error) {
	initCli()
	return openaiCli.ListModels(context.Background())
}

func initCli() {
	once.Do(func() {
		setCli()
	})

	once.Do(func() {
		go func() {
			for {
				select {
				case <-changeSignal:
					setCli()
				}
			}
		}()
	})
}

func setCli() {
	if len(openaiKey) == 0 {
		openaiKey = os.Getenv(constant.OPENAI_API_KEY)
	}
	if len(chimeraKey) == 0 {
		chimeraKey = os.Getenv(constant.CHIMERA_KEY)
	}

	//OpenAI client
	{
		if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
			openaiCli = openai.NewClient(openaiKey)
		} else {
			config := openai.DefaultConfig(openaiKey)
			config.HTTPClient.Timeout = time.Minute * 120
			config.BaseURL = "https://open.aiproxy.xyz/v1"
			openaiCli = openai.NewClientWithConfig(config)
		}
	}
	//chimeraCli
	{
		config := openai.DefaultConfig(chimeraKey)
		config.HTTPClient.Timeout = time.Minute * 120
		config.BaseURL = "https://chimeragpt.adventblocks.cc/api/v1"
		chimeraCli = openai.NewClientWithConfig(config)
	}

}

func SetOpenaiKey(key string) {
	openaiKey = key
	go func() {
		changeSignal <- struct{}{}
	}()
}

func SetChimeraKey(key string) {
	chimeraKey = key
	go func() {
		changeSignal <- struct{}{}
	}()
}
