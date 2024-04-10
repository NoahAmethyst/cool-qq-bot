package top_list

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	"time"
)

func LoadCaiXinNews() ([]spider_pb.CaiXinNew, error) {
	_data, err := spider_svc.CaiXinNews()
	data := make([]spider_pb.CaiXinNew, 0, 50)

	currData := make([]spider_pb.CaiXinNew, 0, 50)

	for _, _new := range _data {
		data = append(data, *_new)
		if !SentRecord.CheckSent(_new.GetTitle()) {
			currData = append(currData, *_new)
			SentRecord.Add(_new.GetTitle())
		}
	}

	if len(currData) > 0 {
		CaiXinNewsDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), data)
	}

	CaiXinNewsDailyRecord.Backup()
	SentRecord.Backup()

	return data, err
}
