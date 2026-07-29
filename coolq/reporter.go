package coolq

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/cluster/jarvis_svc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	"github.com/NoahAmethyst/go-cqhttp/util/finance"
	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (bot *CQBot) openReporter(id int64, isGroup bool) {
	var msg *message.SendingMessage
	if bot.state.reportState.exist(id, isGroup) {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已处于开启状态，无需再次开启")}}
	} else {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已开启")}}
		bot.state.reportState.add(id, isGroup)
	}

	if isGroup {
		bot.SendGroupMessage(id, msg)
	} else {
		bot.SendPrivateMessage(id, 0, msg)
	}

}

func (bot *CQBot) closeReporter(id int64, isGroup bool) {
	var msg *message.SendingMessage
	if bot.state.reportState.exist(id, isGroup) {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已关闭")}}
		bot.state.reportState.del(id, isGroup)
	} else {
		msg = &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("定时推送已处于关闭状态，无需再次关闭")}}
	}

	if isGroup {
		bot.SendGroupMessage(id, msg)
	} else {
		bot.SendPrivateMessage(id, 0, msg)
	}
}

func (bot *CQBot) ReportCoinPrice(group int64, elements []message.IMessageElement, isGroup bool) {
	textEle := getTextEle(elements)

	var symbol string
	if textEle != nil {
		symbol, _ = parseSourceText(textEle)
	}

	priceContents := make([]string, 0, 3)

	symbols := make([]string, 0, 3)

	if len(symbol) > 0 {
		symbols = append(symbols, fmt.Sprintf("%sUSDT", symbol))
	} else {
		symbols = finance.Symbols
	}

	for _, _symbol := range symbols {
		coinInfo, err := finance.Get24HPriceInfo(_symbol)
		if err != nil {
			log.Errorf("get %s error:%s", _symbol, err)
			continue
		}
		priceContents = append(priceContents, fmt.Sprintf("\n%s \n价格：%s\n24小时涨跌幅：%s%% \n最高价：%s \n最低价：%s\n",
			coinInfo.Symbol,
			strings.ReplaceAll(coinInfo.LastPrice, "000", ""),
			coinInfo.PriceChangePercent,
			strings.ReplaceAll(coinInfo.HighPrice, "000", ""),
			strings.ReplaceAll(coinInfo.LowPrice, "000", "")))
	}

	var resp string
	if len(priceContents) == 0 {
		resp = "获取币价实时信息失败，请查看日志"
	} else {
		var coinContent strings.Builder
		coinContent.WriteString(fmt.Sprintf("%s 币价实时信息", time.Now().Format("2006-01-02 15:04")))
		for _, _content := range priceContents {
			coinContent.WriteString(_content)
		}
		resp = coinContent.String()
	}

	if isGroup {
		bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
	} else {
		bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
	}
}

func (bot *CQBot) privateWeiboHot(privateMessage *message.PrivateMessage) {

	textEle := getTextEle(privateMessage.Elements)
	if textEle == nil {
		log.Warnf("QQ消息元素 text element 为空")
		return
	}
	params := parseParam(textEle.Content)
	if len(params) >= 2 {
		if indexList := parseIndexList(params); len(indexList) > 0 {
			bot.ReportSpecificWeibo(privateMessage.Sender.Uin, indexList, false)
		} else {
			bot.ReportWeiboHot([]int64{privateMessage.Sender.Uin}, false)
		}

	} else {
		bot.ReportWeiboHot([]int64{privateMessage.Sender.Uin}, false)
	}

}

func (bot *CQBot) groupWeiboHot(groupMessage *message.GroupMessage) {
	textEle := getTextEle(groupMessage.Elements)
	if textEle == nil {
		log.Warnf("QQ消息元素：text element为空")
		return
	}
	params := parseParam(textEle.Content)
	if len(params) >= 2 {
		if indexList := parseIndexList(params); len(indexList) > 0 {
			bot.ReportSpecificWeibo(groupMessage.GroupCode, indexList, true)
		} else {
			bot.ReportWeiboHot([]int64{groupMessage.GroupCode}, true)
		}
	} else {
		bot.ReportWeiboHot([]int64{groupMessage.GroupCode}, true)
	}
}

func parseIndexList(params []string) []int {
	_indexList := strings.TrimSpace(params[1])
	s1 := strings.Split(_indexList, ",")
	s2 := strings.Split(_indexList, "，")
	indexSet := make(map[int]struct{})
	for _, _index := range s1 {
		if strings.Contains(_index, "，") {
			continue
		}
		index, err := strconv.Atoi(_index)
		if err != nil {
			log.Warnf("parse index %s failed:%s", _index, err.Error())
			continue
		}
		if index-1 < 0 {
			index = 0
		} else {
			index--
		}
		indexSet[index] = struct{}{}
	}

	for _, _index := range s2 {
		if strings.Contains(_index, ",") {
			continue
		}
		index, err := strconv.Atoi(_index)
		if err != nil {
			log.Warnf("parse index %s failed:%s", _index, err.Error())
			continue
		}
		if index-1 < 0 {
			index = 0
		} else {
			index--
		}
		indexSet[index] = struct{}{}
	}

	indexList := make([]int, 0, len(indexSet))
	for _index := range indexSet {
		indexList = append(indexList, _index)
	}

	if len(indexList) > 1 {
		sort.Ints(indexList)
	}

	return indexList
}

func parseParam(content string) []string {
	re := regexp.MustCompile(`#(\S+)\s+(?s)(.*)`)

	match := re.FindStringSubmatch(content)

	params := make([]string, 0, 2)

	if len(match) > 1 {
		params = match[1:]
	}

	return params
}

func (bot *CQBot) ReportSpecificWeibo(group int64, indexList []int, isGroup bool) {
	layout := "2006-01-02 15:04"
	data := top_list.WeiboHotDailyRecord.GetData()
	var lastestT time.Time
	for t := range data {
		reportTime, _ := time.Parse(layout, t)
		if lastestT.Before(reportTime) {
			lastestT = reportTime
		}
	}

	k := lastestT.Format(layout)
	if _data, ok := data[k]; !ok {
		log.Warnf("can't get latest weibo daily report")
		bot.ReportWeiboHot([]int64{group}, isGroup)
	} else {
		for _, _index := range indexList {
			content := fmt.Sprintf("微博热搜#%d\n%s\n链接:%s", _index+1, _data[_index].Title, _data[_index].Url)
			if isGroup {
				bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			} else {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (bot *CQBot) ReportWeiboHot(groups []int64, isGroup bool) {
	var resp string
	if hotList, err := top_list.LoadWeiboHot(); err != nil {
		log.Errorf("get hot list error:%s", err.Error())
		resp = fmt.Sprintf("爬取微博热搜失败：%s", err.Error())

	} else {
		var hotContent strings.Builder
		hotContent.WriteString(fmt.Sprintf("%s 微博实时热搜\n", time.Now().Format("2006-01-02 15:04")))
		for _, hot := range hotList {
			hotContent.WriteString(fmt.Sprintf("%d\t%s\n", hot.Rank, hot.Title))
		}
		resp = hotContent.String()
	}
	for _, _group := range groups {
		if isGroup {
			bot.SendGroupMessage(_group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		} else {
			bot.SendPrivateMessage(_group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		}
	}
}

func (bot *CQBot) Report36kr(groups []int64, isGroup bool) {

	var resp string
	if hotList, err := top_list.Load36krHot(); err != nil {
		log.Errorf("get hot list error:%s", err.Error())
		resp = fmt.Sprintf("爬取36氪热榜失败：%s", err.Error())
	} else {
		var hotContent strings.Builder
		hotContent.WriteString(fmt.Sprintf("%s 36氪24H热榜\n", time.Now().Format("2006-01-02 15:04")))
		for _i, _hot := range hotList {
			if _i > 10 {
				break
			}
			hotContent.WriteString(fmt.Sprintf("%d	%s\n%s\n\n", _hot.Rank, _hot.Title, _hot.Url))
		}
		resp = hotContent.String()
	}
	for _, _group := range groups {
		if isGroup {
			bot.SendGroupMessage(_group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		} else {
			bot.SendPrivateMessage(_group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(resp)}})
		}
	}

}

func (bot *CQBot) ReportWallStreetNews(groups []int64, isGroup bool) bool {
	hasNews := true
	var hasNewsMu sync.Mutex
	hotList, updates, err := top_list.LoadWallStreetNewsWithUpdates()
	if err != nil {
		log.Errorf("爬取华尔街见闻最新资讯失败：%s", err.Error())
	} else {
		go ingestWallStreetNewsUpdates(updates)
		var wait sync.WaitGroup
		for _, _group := range groups {
			wait.Add(1)
			go func(group int64) {
				defer wait.Done()
				readyData := make([]top_list.WallStreetNews, 0, 15)
				for _index, _data := range hotList {
					if _index > 15 {
						break
					}
					if !bot.state.sentNews.checkSent(group, _data.Title) {
						readyData = append(readyData, _data)
					}
				}

				if len(readyData) == 0 {
					//log.Warn().Msgf("华尔街见闻：没有最新资讯，爬取资讯数量:%d", len(hotList))
					hasNewsMu.Lock()
					hasNews = false
					hasNewsMu.Unlock()
				} else {
					//倒序输出，因为最新资讯在第一个
					for i := len(readyData) - 1; i >= 0; i-- {
						bot.state.sentNews.add(group, readyData[i].Title)
						content := fmt.Sprintf("%s\n\n摘要：%s\n\n链接：%s", readyData[i].Title, readyData[i].Content, readyData[i].Url)
						if isGroup {
							bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
						} else {
							bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
						}
						time.Sleep(1 * time.Second)
					}
				}
			}(_group)
		}
		wait.Wait()
		bot.state.sentNews.SaveCache()

	}
	return hasNews
}

func (bot *CQBot) ReportCaiXinNews(groups []int64, isGroup bool) bool {
	hasNews := true
	if news, err := top_list.LoadCaiXinNews(); err != nil {
		log.Errorf("爬取财新网最新新闻失败：%s", err.Error())
	} else {
		var wait sync.WaitGroup
		for _, _group := range groups {
			wait.Add(1)
			go func(group int64) {
				defer wait.Done()
				readyData := make([]spider_pb.CaiXinNew, 0, 15)
				for _index, _data := range news {
					if _index > 15 {
						break
					}
					if !bot.state.sentNews.checkSent(group, _data.GetTitle()) {
						readyData = append(readyData, *_data)
					}
				}

				if len(readyData) == 0 {
					hasNews = false
				} else {
					//倒序输出，因为最新新闻在第一个
					for i := len(readyData) - 1; i >= 0; i-- {
						bot.state.sentNews.add(group, readyData[i].Title)
						content := fmt.Sprintf("【%s】%s\n\n摘要：%s\n\n链接：%s", readyData[i].Domain, readyData[i].Title, readyData[i].Description, readyData[i].Url)
						if isGroup {
							bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
						} else {
							bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
						}
						time.Sleep(1 * time.Second)
					}
				}
			}(_group)
		}
		wait.Wait()
		bot.state.sentNews.SaveCache()

	}
	return hasNews
}

func (bot *CQBot) ReportZhihuHot(group int64, isGroup bool) {
	if hotList, err := top_list.LoadZhihuHot(); err != nil {
		log.Errorf("拉取知乎热榜失败：%s", err.Error())
	} else {
		for i := 0; i < len(hotList); i++ {
			content := fmt.Sprintf("知乎热榜#%d %s\n\n摘要：%s\n\n链接：%s", hotList[i].Rank, hotList[i].Title, hotList[i].Excerpt, hotList[i].Url)
			if isGroup {
				bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			} else {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			}
			time.Sleep(1 * time.Second)
		}
	}
}

// ReportSummary 生成一次每日新闻总结，并推送给全部群聊和私聊订阅者。
func (bot *CQBot) ReportSummary(groupIDs, privateIDs []int64) {
	if len(groupIDs) == 0 && len(privateIDs) == 0 {
		return
	}

	wallStreetData := top_list.WallStreetNewsDailyRecord.GetData()
	if len(wallStreetData) == 0 {
		log.Warnf("没有华尔街今日资讯缓存")
		bot.sendWallStreetSummaryFailure(groupIDs, privateIDs)
		return
	}

	prompt, err := buildWallStreetSummaryPrompt(wallStreetData)
	if err != nil {
		log.Warnf("Jarvis summary preparation failed: operation=build_prompt code=%s", status.Code(err))
		bot.sendWallStreetSummaryFailure(groupIDs, privateIDs)
		return
	}

	response, err := jarvis_svc.Summary(prompt)
	if err != nil {
		log.Errorf("Jarvis summary failed: operation=summary code=%s", status.Code(err))
		bot.sendWallStreetSummaryFailure(groupIDs, privateIDs)
		return
	}
	if response == nil || strings.TrimSpace(response.GetAnswer()) == "" {
		log.Errorf("Jarvis summary failed: operation=validate_response code=%s", codes.Unknown)
		bot.sendWallStreetSummaryFailure(groupIDs, privateIDs)
		return
	}

	chapters, err := parseJarvisSummary(response.GetAnswer())
	if err != nil {
		log.Errorf("Jarvis summary failed: operation=parse_chapters code=%s", status.Code(err))
		bot.sendWallStreetSummaryFailure(groupIDs, privateIDs)
		return
	}

	var wait sync.WaitGroup
	sendToRecipients := func(ids []int64, isGroup bool) {
		for _, reportID := range ids {
			wait.Add(1)
			go func(id int64) {
				defer wait.Done()
				if response.GetLowConfidence() {
					bot.sendWallStreetSummaryMessage(id, isGroup, "提示：Jarvis 将本次总结标记为低置信度，请结合原始新闻核验。")
					time.Sleep(time.Second)
				}
				for chapterIndex, chapter := range chapters {
					for _, chapterMessage := range formatJarvisSummaryChapter(chapterIndex, chapter) {
						bot.sendWallStreetSummaryMessage(id, isGroup, chapterMessage)
						time.Sleep(time.Second)
					}
				}
			}(reportID)
		}
	}
	sendToRecipients(groupIDs, true)
	sendToRecipients(privateIDs, false)
	wait.Wait()
}

func ingestWallStreetNewsUpdates(updates []top_list.WallStreetNews) {
	seenURLs := make(map[string]struct{}, len(updates))
	for _, news := range updates {
		if !isTrustedWallStreetURL(news.Url) {
			continue
		}
		if _, exists := seenURLs[news.Url]; exists {
			continue
		}
		seenURLs[news.Url] = struct{}{}

		content := fmt.Sprintf(
			"新闻日期：%s\n标题：%s\n摘要：%s",
			wallStreetNewsDate(news.CollectedAt),
			news.Title,
			news.Content,
		)
		response, err := jarvis_svc.Ingest(content, news.Url)
		if err != nil {
			log.Errorf("Jarvis Ingest failed: url=%s code=%s", news.Url, status.Code(err))
			continue
		}
		if response == nil || !response.GetSuccess() {
			responseErr := status.Error(codes.Unknown, "ingest rejected")
			log.Errorf("Jarvis Ingest failed: url=%s code=%s", news.Url, status.Code(responseErr))
		}
	}
}

func wallStreetNewsDate(collectedAt string) string {
	shanghai := time.FixedZone("Asia/Shanghai", 8*60*60)
	if collectedTime, err := time.Parse(time.RFC3339, collectedAt); err == nil {
		return collectedTime.In(shanghai).Format("2006-01-02")
	}
	return time.Now().In(shanghai).Format("2006-01-02")
}

func (bot *CQBot) sendWallStreetSummaryFailure(groupIDs, privateIDs []int64) {
	const failureMessage = "华尔街每日资讯总结失败，请稍后查看"
	for _, id := range groupIDs {
		bot.sendWallStreetSummaryMessage(id, true, failureMessage)
	}
	for _, id := range privateIDs {
		bot.sendWallStreetSummaryMessage(id, false, failureMessage)
	}
}

func (bot *CQBot) sendWallStreetSummaryMessage(id int64, isGroup bool, content string) {
	sendingMessage := &message.SendingMessage{
		Elements: []message.IMessageElement{message.NewText(content)},
	}
	if isGroup {
		bot.SendGroupMessage(id, sendingMessage)
		return
	}
	bot.SendPrivateMessage(id, 0, sendingMessage)
}

func goldPriceFluctuation(lastPrice, latestPrice float64) (float64, bool) {
	if lastPrice <= 0 || latestPrice <= 0 {
		return 0, false
	}
	percentage := (latestPrice - lastPrice) / lastPrice * 100
	const float32PercentageTolerance = 0.0001
	return percentage, math.Abs(percentage)+float32PercentageTolerance >= 10
}

type goldPriceAlertSender interface {
	SendGroupMessage(int64, *message.SendingMessage) int32
	SendPrivateMessage(int64, int64, *message.SendingMessage) int32
}

func sendGoldPriceAlert(sender goldPriceAlertSender, groupIDs, privateIDs []int64, content string) {
	for _, id := range groupIDs {
		sender.SendGroupMessage(id, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
	}
	for _, id := range privateIDs {
		sender.SendPrivateMessage(id, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
	}
}

func (bot *CQBot) reportGoldPriceFluctuation(
	groupIDs, privateIDs []int64,
	sender goldPriceAlertSender,
) {
	const goldPriceKey = "last_gold_price"

	price, err := finance.GetGoldPrice()
	if err != nil {
		log.Errorf("获取最新金价失败:%s", err.Error())
		return
	}
	if price <= 0 || math.IsNaN(float64(price)) || math.IsInf(float64(price), 0) {
		log.Errorf("获取最新金价失败:金价必须大于0，实际为%.2f", price)
		return
	}

	currentPrice := float64(price)
	currentPriceText := strconv.FormatFloat(currentPrice, 'f', -1, 32)
	lastPriceText, ok := bot.state.globalState.GetData(goldPriceKey)
	if !ok {
		log.Warn("No last gold price found in bot global state")
		bot.state.globalState.Set(goldPriceKey, currentPriceText)
		log.Infof("Save gold price【%.2f】to bot global state.", price)
		return
	}

	lastPrice, err := strconv.ParseFloat(lastPriceText, 64)
	if err != nil || lastPrice <= 0 || math.IsNaN(lastPrice) || math.IsInf(lastPrice, 0) {
		log.Warnf("Invalid last gold price %q: %v; reset baseline to %.2f", lastPriceText, err, price)
		bot.state.globalState.Set(goldPriceKey, currentPriceText)
		return
	}

	percentage, shouldAlert := goldPriceFluctuation(lastPrice, currentPrice)
	bot.state.globalState.Set(goldPriceKey, currentPriceText)
	log.Infof("Update gold price【%.2f】to bot global state.", price)
	if !shouldAlert {
		log.Infof(
			"last gold price is %.2f, latest gold price is %.2f, fluctuation is %.2f%% under 10%%",
			lastPrice,
			price,
			percentage,
		)
		return
	}

	direction := "上涨"
	if percentage < 0 {
		direction = "下跌"
	}
	content := fmt.Sprintf(
		"金价告警：最新金价为%.2f，相比上一次记录价格（%.2f）%s%.2f%%，请注意风险",
		price,
		lastPrice,
		direction,
		math.Abs(percentage),
	)
	sendGoldPriceAlert(sender, groupIDs, privateIDs, content)
}

func (bot *CQBot) ReportGoldPriceFluctuation(groupIDs, privateIDs []int64) {
	if !bot.goldReportMu.TryLock() {
		log.Warn("Skip overlapping gold price report")
		return
	}
	defer bot.goldReportMu.Unlock()

	bot.reportGoldPriceFluctuation(groupIDs, privateIDs, bot)
}

func getTextEle(eleList []message.IMessageElement) *message.TextElement {
	var textEle *message.TextElement
	for _, _ele := range eleList {
		if v, ok := _ele.(*message.TextElement); ok {
			textEle = v
		}
	}
	return textEle
}
