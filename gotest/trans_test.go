package gotest

import (
	"github.com/Mrs4s/go-cqhttp/util/trans"
	translator_engine "github.com/NoahAmethyst/translator-engine"
	"testing"
)

func Test_Trans(t *testing.T) {

	for i := 0; i < 10; i++ {
		text, err := trans.BalanceTranText("你好", translator_engine.AUTO, translator_engine.EN)
		if err != nil {
			t.Error(err.Error())
		} else {
			t.Logf("%s", text)
		}
	}
}
