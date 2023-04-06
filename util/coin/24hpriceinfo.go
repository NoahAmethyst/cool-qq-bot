package coin

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/http_util"
	"github.com/rs/zerolog/log"
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
	url := fmt.Sprintf("https://api1.binance.com/api/v3/ticker/24hr?symbol=%s", symbol)
	headers := map[string]string{"Accepts": "application/json"}

	if len(os.Getenv(constant.NOT_MIRROR)) == 0 {
		headers["remote"] = "api1.binance.com"
		url = fmt.Sprintf("https://%s/api/v3/ticker/24hr?symbol=%s", os.Getenv(constant.REMOTE_PROXY), symbol)
	}

	err := http_util.GetJSON(url, headers, &data)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request coin pricee", "error": err.Error()}).Send()
	}

	return data, err
}

func init() {
	Symbols = []string{ETH, BTC, BNB}
}
