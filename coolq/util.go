package coolq

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/cluster/spider_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	"github.com/NoahAmethyst/go-cqhttp/util/finance"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func (bot *CQBot) kellyStrategyForPrivate(privateMessage *message.PrivateMessage) {
	response := kellyStrategy(privateMessage.Elements)
	bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
		&message.SendingMessage{Elements: []message.IMessageElement{
			message.NewText(response)}})
}

func (bot *CQBot) kellyStrategyForGroup(groupMessage *message.GroupMessage) {
	response := kellyStrategy(groupMessage.Elements)
	bot.SendGroupMessage(groupMessage.GroupCode,
		&message.SendingMessage{Elements: []message.IMessageElement{
			message.NewReply(groupMessage),
			message.NewText(response)}})
}

func kellyStrategy(elements []message.IMessageElement) string {
	text := getTextEle(elements)
	numbers, err := parseAllNumber(text.Content)
	response := ""
	if err != nil || len(numbers) != 3 {
		response = "要使用凯利公式，请依次输入【潜在正收益率】、【潜在损失率】、【收益概率/获胜概率】，输入数值为概率x100，本公式将返回依据凯利公式计算的本次投注金额占总金额比例"
		return response
	}
	if fStar, _err := calculateKelly(numbers[0], numbers[1], numbers[2]); _err != nil {
		response = _err.Error()
	} else {
		if fStar > 0 {
			response = fmt.Sprintf("依据凯利公式(Kelly Strategy)，本次策略中投资金额占总金额的【%.2f%%】", fStar*100)
		} else {
			response = fmt.Sprintf("依据凯利公式(Kelly Strategy)，应该放弃这次投资")
		}

	}
	return response
}

func parseAllNumber(text string) ([]float64, error) {
	re := regexp.MustCompile(`-?\d+(\.\d+)?`) // \d+ 匹配一个或多个数字
	// 使用 FindAllString 查找所有匹配的数字
	_numbers := re.FindAllString(text, -1) // -1 表示返回
	numbers := make([]float64, 0, 3)
	for _, number := range _numbers {
		float, err := strconv.ParseFloat(number, 10)
		if err != nil {
			return numbers, err
		}
		numbers = append(numbers, float)
	}
	return numbers, nil
}

func calculateKelly(b, l, p float64) (float64, error) {
	q := 1 - p/100 // 失败的概率

	// make sure that b > 0 and 0 <= p <= 1
	if b <= 0 || l <= 0 || p < 0 || p > 100 {
		return 0, errors.New("【潜在收益率】、【潜在损失率】以及【收益概率/获胜概率】必须大于0，且【收益概率/获胜概率】必须不大于100")
	}
	// use kelly strategy f* = (b * p - q) / b
	fStar := ((b/100)*p - (l/100)*q) / ((b / 100) * (l * 100))
	return fStar, nil
}

func (bot *CQBot) goldPriceForPrivate(privateMessage *message.PrivateMessage) {
	if response, err := spider_svc.Finance(spider_pb.FinanceType_GOLD, "", ""); err != nil {
		log.Errorf("获取最新金价失败：%s", err.Error())
		bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("获取最新金价失败，请查看日志")}})
	} else {
		bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText(fmt.Sprintf("今日最新金价：%.2f", response.FloatValue))}})
	}

}

func (bot *CQBot) goldPriceForGroup(groupMessage *message.GroupMessage) {
	if _price, err := finance.GetGoldPrice(); err != nil {
		log.Errorf("获取最新金价失败：%s", err.Error())
		bot.SendGroupMessage(groupMessage.GroupCode,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("获取最新金价失败，请查看日志")}})
	} else {
		bot.SendGroupMessage(groupMessage.GroupCode,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewReply(groupMessage),
				message.NewText(fmt.Sprintf("今日最新金价：%.2f", _price))}})
	}
}

func (bot *CQBot) exChangeRateForPrivate(privateMessage *message.PrivateMessage) {
	text := getTextEle(privateMessage.Elements)
	_from, _to, err := extractCurrenciesRegex(text.Content)
	if err != nil {
		bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText(err.Error())}})
		return
	}

	if _rate, err := finance.ExchangeRate(_from, _to); err != nil {
		log.Errorf("获取最新汇率失败：%s", err.Error())
		bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("获取汇率失败，请查看日志")}})
	} else {
		bot.SendPrivateMessage(privateMessage.Sender.Uin, 0,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText(fmt.Sprintf("【%s】与【%s】的汇率为：%.3f", _from, _to, _rate))}})
	}
}

func (bot *CQBot) exChangeRateForGroup(groupMessage *message.GroupMessage) {
	text := getTextEle(groupMessage.Elements)
	_from, _to, err := extractCurrenciesRegex(text.Content)
	if err != nil {
		bot.SendGroupMessage(groupMessage.GroupCode,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText(err.Error())}})
		return
	}

	if _rate, err := finance.ExchangeRate(_from, _to); err != nil {
		log.Errorf("获取最新汇率失败：%s", err.Error())
		bot.SendGroupMessage(groupMessage.GroupCode,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText("获取汇率失败，请查看日志")}})
	} else {
		bot.SendGroupMessage(groupMessage.GroupCode,
			&message.SendingMessage{Elements: []message.IMessageElement{
				message.NewText(fmt.Sprintf("【%s】与【%s】的汇率为：%.3f", _from, _to, _rate))}})
	}
}

func extractCurrenciesRegex(text string) (string, string, error) {
	re := regexp.MustCompile(`#汇率\s+(\S+)\s+(\S+)`)
	matches := re.FindStringSubmatch(text)

	if len(matches) != 3 {
		log.Errorf("汇率格式错误：%s", text)
		var _err_message strings.Builder
		if _currency_list, err := finance.SupportCurrencies(); err != nil {
			_err_message.WriteString("汇率命令格式错误，举例：\"#汇率 美金 人民币\"")
		} else {
			_err_message.WriteString("汇率命令格式错误，举例：\"#汇率 美金 人民币\",支持的币种列表：")
			for _, _currency := range _currency_list {
				_err_message.WriteString(_currency)
				_err_message.WriteString(" ")
			}
		}
		return "", "", fmt.Errorf(_err_message.String())
	}

	return matches[1], matches[2], nil
}
