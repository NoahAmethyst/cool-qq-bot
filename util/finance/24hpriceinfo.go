package finance

import (
	"fmt"
	"github.com/NoahAmethyst/go-cqhttp/constant"
	"github.com/NoahAmethyst/go-cqhttp/util/http_util"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	ETH = "ETHUSDT"
	BTC = "BTCUSDT"
	BNB = "BNBUSDT"
)

var Symbols []string

type CoinPrice struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
}

func Get24HPriceInfo(symbol string) (CoinPrice, error) {

	data := CoinPrice{}
	url := fmt.Sprintf("https://data.binance.com/api/v3/ticker/24hr?symbol=%s", symbol)
	headers := map[string]string{"Accepts": "application/json"}

	if len(os.Getenv(constant.NOT_MIRROR)) == 0 {
		headers["remote"] = "data.binance.com"
		url = fmt.Sprintf("https://%s/api/v3/ticker/24hr?symbol=%s", os.Getenv(constant.REMOTE_PROXY), symbol)
	}

	err := http_util.GetJSON(url, headers, &data)
	if err != nil {
		log.Errorf("request coin price failed:%s", err.Error())
	}

	return data, err
}

func init() {
	Symbols = []string{ETH, BTC, BNB}
}
