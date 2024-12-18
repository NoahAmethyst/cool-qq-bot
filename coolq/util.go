package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
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
