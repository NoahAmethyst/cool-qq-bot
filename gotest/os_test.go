package gotest

import (
	"regexp"
	"testing"
	"time"
)

func TestCommon(t *testing.T) {
	t.Logf("time:%s", time.Now().Format("20060102"))
}

func TestParseNumber(t *testing.T) {

	text := "10 20.6,30 40"
	re := regexp.MustCompile(`-?\d+(\.\d+)?`) // \d+ 匹配一个或多个数字
	// 使用 FindAllString 查找所有匹配的数字
	numbers := re.FindAllString(text, -1) // -1 表示返回

	t.Logf("%+v", numbers)
}
