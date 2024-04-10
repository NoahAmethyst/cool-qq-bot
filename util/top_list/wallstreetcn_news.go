package top_list

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
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
	currData := make([]WallStreetNews, 0, 50)
	for _, _new := range _news {
		_d := WallStreetNews{
			Title:   _new.Title,
			Content: _new.Content,
			Url:     _new.Url,
		}
		data = append(data, _d)
		if !SentRecord.CheckSent(_new.GetTitle()) {
			currData = append(currData, _d)
			SentRecord.Add(_new.GetTitle())
		}
	}

	if len(currData) > 0 {
		WallStreetNewsDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), currData)
	}

	WallStreetNewsDailyRecord.Backup()
	SentRecord.Backup()

	return data, err
}
