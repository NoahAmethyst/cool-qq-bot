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

func TestKelly(t *testing.T) {
	b := 10
	p := 20

	q := 100 - p // 失败的概率

	// make sure that b > 0 and 0 <= p <= 1
	// use kelly strategy f* = (b * p - q) / b
	fStar := (b*p - q) / b
	t.Logf("%+v", fStar)

}
