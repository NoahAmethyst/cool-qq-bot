package top_list

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	_ "github.com/NoahAmethyst/go-cqhttp/util/file_util"
	"strconv"
	"time"
)

type WeiboHot struct {
	Title string
	Hot   string
	Rank  int
	Url   string
}

func LoadWeiboHot() ([]WeiboHot, error) {
	_hotList, err := spider_svc.WeiboHot()
	hotList := make([]WeiboHot, 0, 50)
	for _, _hot := range _hotList {
		hotList = append(hotList, WeiboHot{
			Title: _hot.Title,
			Url:   _hot.Url,
			Hot:   strconv.Itoa(int(_hot.Hot)),
			Rank:  int(_hot.Rank),
		})
	}

	WeiboHotDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), hotList)

	WeiboHotDailyRecord.Backup()

	return hotList, err
}
