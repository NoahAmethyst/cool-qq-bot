package top_list

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	_ "github.com/NoahAmethyst/go-cqhttp/util/file_util"
	"time"
)

type Data36krHot struct {
	Title string
	Url   string
	Rank  int
}

func Load36krHot() ([]Data36krHot, error) {
	_data, err := spider_svc.D36Kr()
	data := make([]Data36krHot, 0, 20)
	for _, _d := range _data {
		data = append(data, Data36krHot{
			Title: _d.Title,
			Url:   _d.Url,
			Rank:  int(_d.Rank),
		})
	}
	D36krDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), data)
	D36krDailyRecord.Backup()

	return data, err
}
