package gotest

import (
	"github.com/NoahAmethyst/go-cqhttp/util/encrypt"
	"testing"
)

func Test_Hash(t *testing.T) {
	s1 := "恒指高开0.54%，恒生科技指数涨1.71%"
	s2 := "恒指高开0.54%，恒生科技指数涨1.71%"
	hash1 := encrypt.HashStr(s1)
	hash2 := encrypt.HashStr(s2)
	t.Logf("%v", hash1 == hash2)
}
