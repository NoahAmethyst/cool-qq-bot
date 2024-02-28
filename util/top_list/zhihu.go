package top_list

import (
	"github.com/Mrs4s/go-cqhttp/cluster/spider_svc"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"time"
)

type ZhihuHot struct {
	Title   string
	Url     string
	Excerpt string
	Rank    int
	Created int64
}

func LoadZhihuHot() ([]ZhihuHot, error) {
	_hotList, err := spider_svc.ZhihuHot()

	hotList := make([]ZhihuHot, 0, 50)
	for _, _hot := range _hotList {
		hotList = append(hotList, ZhihuHot{
			Title:   _hot.Title,
			Url:     _hot.Url,
			Excerpt: _hot.Excerpt,
			Rank:    int(_hot.Rank),
			Created: _hot.Created,
		})
	}

	ZhihuHotDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), hotList)

	go func() {
		path := file_util.GetFileRoot()
		_, _ = file_util.WriteJsonFile(ZhihuHotDailyRecord.GetData(), path, "zhihu", true)
	}()

	//only return top 10 of the hot list
	return hotList[:10], err
}
