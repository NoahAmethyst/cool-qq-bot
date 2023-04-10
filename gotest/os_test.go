package gotest

import (
	"testing"
	"time"
)

func TestCommon(t *testing.T) {
	t.Logf("time:%s", time.Now().Format("20060102"))
}
