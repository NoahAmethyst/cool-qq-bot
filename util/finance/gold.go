package finance

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	log "github.com/sirupsen/logrus"
)

func GetGoldPrice() (float32, error) {
	resp, err := spider_svc.Finance(spider_pb.FinanceType_GOLD, "", "")
	if err != nil {
		log.Errorf("Get gold price error:%s", err.Error())
	}
	return resp.FloatValue, err
}
