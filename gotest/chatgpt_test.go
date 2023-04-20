package gotest

import (
	"github.com/Mrs4s/go-cqhttp/util/openai_util"
	"testing"
)

func Test_Chatgpt(t *testing.T) {
	replyMsg, err := openai_util.AskChatGpt("hello")
	if err != nil {
		panic(err)
	}
	t.Logf("%s", replyMsg)
}

func Test_AIAssistant(t *testing.T) {
	assistant, err := openai_util.AskAIAssistant("1+1=?")
	if err != nil {
		panic(err)
	}
	t.Logf("%s", assistant.Text)

	assistant, err = openai_util.AskAIAssistant("这个结果的二次方等于多少", assistant.ID)
	if err != nil {
		panic(err)
	}
	t.Logf("%s", assistant.Text)
}
