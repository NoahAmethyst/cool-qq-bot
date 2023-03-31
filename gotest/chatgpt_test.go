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
