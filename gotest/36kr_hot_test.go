package gotest

import (
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"testing"
)

func Test_36kr(t *testing.T) {

	data, _ := top_list.Load36krHot()

	for _, _data := range data {
		t.Logf("%+v", _data)
	}
}
