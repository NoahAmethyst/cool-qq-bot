package finance

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	log "github.com/sirupsen/logrus"
)

func GetGoldPrice() (float32, error) {
	goldPrice := float32(0)
	resp, err := spider_svc.Finance(spider_pb.FinanceType_GOLD, "", "")
	if err != nil {
		log.Errorf("Get gold price error:%s", err.Error())
	}
	if resp == nil {
		log.Errorf("Get nil resp from spider svc")
	} else {
		goldPrice = resp.FloatValue
	}
	return goldPrice, err
}
