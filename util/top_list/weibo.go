package top_list

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	_ "github.com/NoahAmethyst/go-cqhttp/util/file_util"
	"strconv"
	"time"
)

type WeiboHot struct {
	Title string
	Hot   string
	Rank  int
	Url   string
	Tags  []string
}

func LoadWeiboHot() ([]WeiboHot, error) {
	_hotList, err := spider_svc.WeiboHot()
	hotList := make([]WeiboHot, 0, 50)
	for _, _hot := range _hotList {
		hotList = append(hotList, weiboHotFromProto(_hot))
	}

	WeiboHotDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), hotList)

	WeiboHotDailyRecord.Backup()

	return hotList, err
}

func weiboHotFromProto(hot *spider_pb.WeiboHot) WeiboHot {
	return WeiboHot{
		Title: hot.GetTitle(),
		Url:   hot.GetUrl(),
		Hot:   strconv.Itoa(int(hot.GetHot())),
		Rank:  int(hot.GetRank()),
		Tags:  append([]string(nil), hot.GetTags()...),
	}
}
