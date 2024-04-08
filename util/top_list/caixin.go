package top_list

import "C"
import (
	"github.com/Mrs4s/go-cqhttp/cluster/spider_svc"
	"github.com/Mrs4s/go-cqhttp/protocol/pb/spider_pb"
	"github.com/Mrs4s/go-cqhttp/util/encrypt"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"time"
)

func LoadCaiXinNews() ([]spider_pb.CaiXinNew, error) {
	_data, err := spider_svc.CaiXinNews()
	data := make([]spider_pb.CaiXinNew, 0, 50)

	currData := make([]spider_pb.CaiXinNew, 0, 50)

	for _, _new := range _data {
		data = append(data, *_new)
	}

	if recordData := CaiXinNewsDailyRecord.GetData(); len(recordData) > 0 {
		restoredSet := make(map[uint32]struct{})
		for _, _news := range recordData {
			for _, _existNew := range _news {
				restoredSet[encrypt.HashStr(_existNew.Title)] = struct{}{}
			}
		}

		for _, _new := range data {
			if _, ok := restoredSet[encrypt.HashStr(_new.Title)]; !ok {
				currData = append(currData, spider_pb.CaiXinNew{
					Title:       _new.Title,
					Description: _new.Description,
					Url:         _new.Url,
					Domain:      _new.Domain,
				})
			}
		}
	} else {
		currData = data
	}

	if len(currData) > 0 {
		CaiXinNewsDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), data)
	}

	go func() {
		path := file_util.GetFileRoot()
		_, _ = file_util.WriteJsonFile(CaiXinNewsDailyRecord.GetData(), path, "caixin_news", true)
	}()

	return data, err
}
