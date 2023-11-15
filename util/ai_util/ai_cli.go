package ai_util

import (
	"context"
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	go_ernie "github.com/anhao/go-ernie"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

var openaiCli *openai.Client
var chimeraCli *openai.Client
var ernieCli *go_ernie.Client

var openaiKey string
var chimeraKey string
var ernieKey string
var ernieSecret string

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
		ernieKey = os.Getenv(constant.ERNIE_APP_KEY)
		ernieSecret = os.Getenv(constant.ERNIE_APP_SECRET)
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
