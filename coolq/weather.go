package coolq

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/cluster/jarvis_svc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (bot *CQBot) handlePrivateWeather(privateMessage *message.PrivateMessage) {
	userID := privateMessage.Sender.Uin
	textEle := getTextEle(privateMessage.Elements)
	if textEle == nil {
		bot.replyPrivateWeather(userID, "天气设置格式：#天气 城市\n例如：#天气 杭州")
		return
	}

	value, missing := parseSourceText(textEle)
	if missing {
		if city, ok := bot.state.weatherState.get(userID); ok {
			bot.replyPrivateWeather(userID, fmt.Sprintf("当前天气城市：%s\n推送时间：每天 08:00", city))
			return
		}
		bot.replyPrivateWeather(userID, "尚未设置天气城市\n请发送：#天气 城市\n例如：#天气 杭州")
		return
	}

	value = strings.TrimSpace(value)
	if value == "关闭" {
		if bot.state.weatherState.del(userID) {
			bot.replyPrivateWeather(userID, "天气推送已关闭")
			return
		}
		bot.replyPrivateWeather(userID, "天气推送已处于关闭状态")
		return
	}

	if !validWeatherCity(value) {
		bot.replyPrivateWeather(userID, "城市名称无效，请发送不超过 40 个字符的城市名称\n例如：#天气 杭州")
		return
	}

	bot.state.weatherState.set(userID, value)
	bot.replyPrivateWeather(
		userID,
		fmt.Sprintf("天气城市已设置为：%s\n每天 08:00 为你推送天气与出行建议", value),
	)
}

func validWeatherCity(city string) bool {
	if city == "" || utf8.RuneCountInString(city) > 40 {
		return false
	}
	for _, char := range city {
		if unicode.IsControl(char) {
			return false
		}
	}
	return true
}

func (bot *CQBot) replyPrivateWeather(userID int64, content string) {
	bot.SendPrivateMessage(userID, 0, &message.SendingMessage{
		Elements: []message.IMessageElement{message.NewText(content)},
	})
}

func buildWeatherPrompt(city string, now time.Time) string {
	return fmt.Sprintf(
		`今天是%s，时区为Asia/Shanghai。请使用你可用的天气查询能力，查询%s今天的天气预报，并给出适合当天执行的生活建议。

内容要求：
先用简洁文字说明天气状况，并提供能够确认的最高温、最低温、降水概率或降水时段、风力和紫外线信息。然后分别给出今日出行、防晒、穿搭建议和其他注意事项。无法确认的数据必须明确说明信息暂缺，不得猜测或虚构。全文控制在800个汉字以内。

输出格式要求：
返回适合直接发送到QQ的普通文本，可以使用emoji和符号表情增强可读性。不要使用Markdown，不要使用Markdown标题、项目符号、加粗、斜体、删除线、代码块、表格、引用或Markdown链接。可以使用简短自然段、普通中文标点和emoji组织内容。`,
		now.In(time.FixedZone("Asia/Shanghai", 8*60*60)).Format("2006年01月02日"),
		city,
	)
}

func weatherJarvisUserID(city string) string {
	hash := sha256.Sum256([]byte(strings.TrimSpace(city)))
	return fmt.Sprintf("go-cqhttp:weather:%x", hash[:12])
}

func (bot *CQBot) ReportWeather(subscriptions map[int64]string) {
	if len(subscriptions) == 0 {
		return
	}

	recipientsByCity := make(map[string][]int64)
	for userID, city := range subscriptions {
		city = strings.TrimSpace(city)
		if !validWeatherCity(city) {
			continue
		}
		recipientsByCity[city] = append(recipientsByCity[city], userID)
	}

	cities := make([]string, 0, len(recipientsByCity))
	for city := range recipientsByCity {
		cities = append(cities, city)
	}
	sort.Strings(cities)

	for _, city := range cities {
		userIDs := recipientsByCity[city]
		sort.Slice(userIDs, func(i, j int) bool {
			return userIDs[i] < userIDs[j]
		})

		response, err := jarvis_svc.Chat(
			buildWeatherPrompt(city, time.Now()),
			weatherJarvisUserID(city),
		)
		if err != nil {
			log.Errorf("Jarvis weather query failed: operation=chat code=%s", status.Code(err))
			bot.sendWeatherFailure(userIDs)
			continue
		}

		if response == nil {
			log.Errorf("Jarvis weather query failed: operation=validate_response code=%s", codes.Unknown)
			bot.sendWeatherFailure(userIDs)
			continue
		}

		content := toPlainQQText(response.GetAnswer())
		if content == "" {
			log.Errorf("Jarvis weather query failed: operation=sanitize_response code=%s", codes.Unknown)
			bot.sendWeatherFailure(userIDs)
			continue
		}
		if response.GetLowConfidence() {
			content = "提示：本次天气信息置信度较低，请以当地气象部门最新预报为准。\n\n" + content
		}

		for _, userID := range userIDs {
			bot.replyPrivateWeather(userID, content)
		}
	}
}

func (bot *CQBot) sendWeatherFailure(userIDs []int64) {
	for _, userID := range userIDs {
		bot.replyPrivateWeather(userID, "今日天气查询失败，请稍后重试")
	}
}
