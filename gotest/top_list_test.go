package gotest

import (
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"testing"
)

func Test_Zhihu(t *testing.T) {
	data, err := top_list.LoadZhihuHot()
	if err != nil {
		panic(err)
	}
	for _, _data := range data {
		t.Logf("%+v", _data)
	}

}

func Test_36kr(t *testing.T) {

	data, _ := top_list.Load36krHot()

	for _, _data := range data {
		t.Logf("%+v", _data)
	}
}

func Test_WeiboHot(t *testing.T) {
	hotList, err := top_list.LoadWeiboHot()
	if err != nil {
		t.Error(err)
	} else {
		for _, _hot := range hotList {
			t.Logf("%+v", _hot)
		}
	}
}

func Test_WallStreetcn_Latest(t *testing.T) {
	data, err := top_list.LoadWallStreetNews()
	if err != nil {
		panic(err)
	}

	for _, _data := range data {
		t.Logf("%+v", _data)
	}

}
