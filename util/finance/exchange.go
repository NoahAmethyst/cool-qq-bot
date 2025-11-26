package finance

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	log "github.com/sirupsen/logrus"
)

func SupportCurrencies() []string {
	resp, err := spider_svc.Finance(spider_pb.FinanceType_CURRENCY_LIST, "", "")
	if err != nil {
		log.Errorf("Get gold price error:%s", err.Error())
	}
	return resp.StrList
}

func ExchangeRate(_from string, _to string) float32 {
	resp, err := spider_svc.Finance(spider_pb.FinanceType_EXCHANGE, _from, _to)
	if err != nil {
		log.Errorf("Get gold price error:%s", err.Error())
	}
	return resp.FloatValue
}
