package gotest

import (
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"testing"
)

func Test_WallStreetcn_Latest(t *testing.T) {
	data, err := top_list.LoadWallStreetNews()
	if err != nil {
		panic(err)
	}

	for _, _data := range data {
		t.Logf("%+v", _data)
	}

	data, err = top_list.LoadWallStreetNews()
	if err != nil {
		panic(err)
	}

	for _, _data := range data {
		t.Logf("%+v", _data)
	}

}
