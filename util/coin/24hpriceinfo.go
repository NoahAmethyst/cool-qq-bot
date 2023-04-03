package coin

import (
	"fmt"
	"github.com/Mrs4s/go-cqhttp/util/http_util"
	"github.com/rs/zerolog/log"
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
	err := http_util.GetJSON(url, map[string]string{"Accepts": "application/json"}, &data)
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "request coin pricee", "error": err.Error()}).Send()
	}

	return data, err
}

func init() {
	Symbols = []string{ETH, BTC, BNB}
}
