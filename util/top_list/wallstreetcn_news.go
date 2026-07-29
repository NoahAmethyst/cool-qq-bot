package top_list

import (
	"strings"
	"time"

	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
)

type WallStreetNews struct {
	Title       string
	Content     string
	Url         string
	CollectedAt string
}

func LoadWallStreetNews() ([]WallStreetNews, error) {
	data, _, err := LoadWallStreetNewsWithUpdates()
	return data, err
}

func LoadWallStreetNewsWithUpdates() (all []WallStreetNews, updates []WallStreetNews, err error) {
	_news, err := spider_svc.WallStreetNews()
	now := time.Now()
	collectedAt := now.Format(time.RFC3339)

	data := make([]WallStreetNews, 0, 50)
	currData := make([]WallStreetNews, 0, 50)
	for _, _new := range _news {
		_d := WallStreetNews{
			Title:       _new.Title,
			Content:     _new.Content,
			Url:         _new.Url,
			CollectedAt: collectedAt,
		}
		data = append(data, _d)
		if SentRecord.MarkIfNew(wallStreetNewsDedupKey(_d)) {
			currData = append(currData, _d)
		}
	}

	if len(currData) > 0 {
		WallStreetNewsDailyRecord.Add(now.Format("2006-01-02 15:04"), currData)
	}

	WallStreetNewsDailyRecord.Backup()
	SentRecord.Backup()

	return data, currData, err
}

func wallStreetNewsDedupKey(news WallStreetNews) string {
	return strings.Join([]string{news.Url, news.Title, news.Content}, "\x00")
}
