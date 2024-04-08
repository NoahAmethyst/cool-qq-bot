package top_list

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/util/encrypt"
	"github.com/NoahAmethyst/go-cqhttp/util/file_util"
	"time"
)

type WallStreetNews struct {
	Title   string
	Content string
	Url     string
}

func LoadWallStreetNews() ([]WallStreetNews, error) {
	_news, err := spider_svc.WallStreetNews()

	data := make([]WallStreetNews, 0, 50)

	for _, _new := range _news {
		data = append(data, WallStreetNews{
			Title:   _new.Title,
			Content: _new.Content,
			Url:     _new.Url,
		})
	}

	currData := make([]WallStreetNews, 0, 50)

	if recordData := WallStreetNewsDailyRecord.GetData(); len(recordData) > 0 {
		restoredSet := make(map[uint32]struct{})

		for _, news := range recordData {
			for _, _existNew := range news {
				restoredSet[encrypt.HashStr(_existNew.Title)] = struct{}{}
			}
		}
		for _, _data := range data {
			if _, ok := restoredSet[encrypt.HashStr(_data.Title)]; !ok {
				currData = append(currData, WallStreetNews{
					Title:   _data.Title,
					Content: _data.Content,
					Url:     _data.Url,
				})
			}
		}
	} else {
		currData = data
	}

	if len(currData) > 0 {
		WallStreetNewsDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), currData)
	}

	go func() {
		path := file_util.GetFileRoot()
		_, _ = file_util.WriteJsonFile(WallStreetNewsDailyRecord.GetData(), path, "wallstreet_news", true)
	}()

	return data, err
}
