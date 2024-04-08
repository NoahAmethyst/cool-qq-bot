package top_list

import "C"
import (
	"github.com/Mrs4s/go-cqhttp/cluster/spider_svc"
	"github.com/Mrs4s/go-cqhttp/protocol/pb/spider_pb"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"time"
)

func LoadCaiXinNews() ([]spider_pb.CaiXinNew, error) {
	_data, err := spider_svc.CaiXinNews()
	data := make([]spider_pb.CaiXinNew, 0, 50)
	for _, _new := range _data {
		data = append(data, *_new)
	}

	CaiXinNewsDailyRecord.Add(time.Now().Format("2006-01-02 15:04"), data)
	go func() {
		path := file_util.GetFileRoot()
		_, _ = file_util.WriteJsonFile(D36krDailyRecord.GetData(), path, "36kr", true)
	}()

	return data, err
}
