package ai_util

import (
	"context"
	"fmt"
	"github.com/NoahAmethyst/go-cqhttp/constant"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

var AIAssistantAttributions map[ChatModel]AIAssistantAttribution

type AIAssistantAttribution struct {
	Name   string
	Vendor string
}

var openaiCli *openai.Client
var chimeraCli *openai.Client
var ernieCli *go_ernie.Client
var deepSeekCli *openai.Client

var openaiKey string
var chimeraKey string
var ernieKey string
var ernieSecret string
var deepSeekApiKey string

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

	if len(ernieKey) == 0 || len(ernieSecret) == 0 {
		ernieKey = os.Getenv(constant.QIANFAN_ACCESS_KEY)
		ernieSecret = os.Getenv(constant.QIANFAN_SECRET_KEY)
	}

	if len(deepSeekApiKey) == 0 {
		deepSeekApiKey = os.Getenv(constant.DEEPSEEK_API_KEY)
	}
	//OpenAI client
	{
		if len(openaiKey) > 0 {
			log.Info("init OpenAI client")
			if len(os.Getenv(constant.NOT_MIRROR)) > 0 {
				openaiCli = openai.NewClient(openaiKey)
			} else {
				openaiProxy := os.Getenv(constant.OPENAI_PROXY)
				config := openai.DefaultConfig(openaiKey)
				config.HTTPClient.Timeout = time.Minute * 120
				config.BaseURL = fmt.Sprintf("https://%s/v1", openaiProxy)
				openaiCli = openai.NewClientWithConfig(config)
			}
		}

	}
	//chimeraCli
	{
		if len(chimeraKey) > 0 {
			log.Info("init ChiemraGpt client")
			config := openai.DefaultConfig(chimeraKey)
			config.HTTPClient.Timeout = time.Minute * 120
			config.BaseURL = "https://chimeragpt.adventblocks.cc/api/v1"
			chimeraCli = openai.NewClientWithConfig(config)
		}

	}
	//ErnieCli
	{
		if len(ernieKey) > 0 && len(ernieSecret) > 0 {
			log.Info("init Ernie client")
			ernieCli = go_ernie.NewDefaultClient(ernieKey, ernieSecret)
		}
	}
	//DeepSeek
	{
		if len(deepSeekApiKey) > 0 {
			log.Info("init DeepSeek client")
			config := openai.DefaultConfig(deepSeekApiKey)
			config.HTTPClient.Timeout = time.Minute * 120
			config.BaseURL = deepSeekBaseUrl
			deepSeekCli = openai.NewClientWithConfig(config)
		}
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

func init() {
	AIAssistantAttributions = map[ChatModel]AIAssistantAttribution{
		ChatGPT:     {Name: "ChatGPT(3.5)", Vendor: "OpenAI"},
		ChatGPT4:    {Name: "ChatGPT(4.0)", Vendor: "OpenAI"},
		BingCopilot: {Name: "Bing Copilot", Vendor: "Microsoft"},
		Ernie:       {Name: "百度千帆", Vendor: "Baidu"},
		DeepSeek:    {Name: "Deep Seek", Vendor: "Deep Seek"},
	}
}
