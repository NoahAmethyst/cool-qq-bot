package top_list

import (
	"reflect"
	"testing"

	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
)

func TestWeiboHotFromProtoCopiesTags(t *testing.T) {
	want := []string{"当前爆词"}
	got := weiboHotFromProto(&spider_pb.WeiboHot{Tags: want})
	if !reflect.DeepEqual(got.Tags, want) {
		t.Fatalf("Tags = %#v, want %#v", got.Tags, want)
	}
}
