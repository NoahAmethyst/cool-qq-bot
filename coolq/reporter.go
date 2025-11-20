package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
	"github.com/NoahAmethyst/go-cqhttp/util/ai_util"
	"github.com/NoahAmethyst/go-cqhttp/util/coin"
	"github.com/NoahAmethyst/go-cqhttp/util/content_util"
	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
	"github.com/sashabaranov/go-openai"
	log "github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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
		symbols = coin.Symbols
	}

	for _, _symbol := range symbols {
		coinInfo, err := coin.Get24HPriceInfo(_symbol)
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
	if hotList, err := top_list.LoadWallStreetNews(); err != nil {
		log.Errorf("爬取华尔街见闻最新资讯失败：%s", err.Error())
	} else {
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
					hasNews = false
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

// ReportSummary 推送每日新闻总结
func (bot *CQBot) ReportSummary(groups []int64, isGroup bool) {
	// Summary wall street news
	if wallStreetData := top_list.WallStreetNewsDailyRecord.GetData(); len(wallStreetData) == 0 {
		log.Warnf("没有华尔街今日资讯缓存")
	} else {
		var content strings.Builder
		date := time.Now().Format("2006-01-02")
		for k, _list := range wallStreetData {
			if len(_list) == 0 {
				log.Warnf("时间 %s 的华尔街资讯为空", k)
			} else {
				date = k
				for _, _data := range _list {
					content.WriteString(fmt.Sprintf("标题：%s,内容：%s\n", _data.Title, _data.Content))
				}
			}

		}
		if content.Len() == 0 {
			log.Warnf("No wall street news found.")
			return
		}

		content.WriteString(fmt.Sprintf("上面是我提供的新闻内容,新闻日期为%s，根据我提供的今日经济新闻列表（包含标题和概述），生成一份专业的“每日全球经济形势总结报告”。\n报告输出要求与结构：\n\n请务必严格按照以下四个部分的结构和内容要点进行组织：\n\n第一部分：今日全球经济形势概览\n\n核心基调判断： 用一两句话概括今日全球市场呈现出的主要情绪和趋势（例如：风险偏好上升/下降、避险情绪浓厚、增长担忧加剧、通胀预期分化等）。\n\n主要驱动力： 指出是哪些关键事件或数据（来自新闻列表）主导了今日的市场基调。\n\n第二部分：重点事件与影响深度分析\n\n请提炼出今日最具影响力的2-4个核心事件。\n\n对每个核心事件，按以下层次进行分析：\n\n事件简述： 简要说明是什么事。\n\n直接影响： 分析该事件对特定市场（如股市、债市、外汇、大宗商品）、国家/地区或行业的直接影响。\n\n潜在深层影响： 推测该事件可能引发的连锁反应或长期政策含义。\n\n第三部分：关键关联性推测\n\n这是报告的核心。请分析不同新闻事件之间存在的逻辑关联，并给出合理解释。\n\n跨市场关联： 例如，A国的货币政策新闻如何影响到B国的债券市场，并进而波及C国的大宗商品价格。\n\n因果与传导链条： 尝试构建一个或多个“因为事件A，所以可能导致结果B，进而需要关注C”的逻辑链条。\n\n主题关联： 将看似孤立的事件，归纳到更大的主题下（例如：“全球央行政策分化”、“能源转型的地缘政治影响”、“供应链重构新动向”）。\n\n第四部分：未来展望与风险提示\n\n明日/近期关注点： 基于今日事件的发展，指出未来1-3个交易日需要重点关注的数据、事件或政策信号。\n\n潜在风险预警： 提示当前市场中可能被忽视或正在酝酿的下行风险。\n\n机会提示： 简要提及今日事件可能带来的潜在投资或市场机会。\n\n语言与风格：\n\n专业、严谨、客观。\n\n使用规范的宏观经济术语，但对可能难以理解的概念可稍作解释。\n\n避免使用“可能”、“也许”等模糊词汇，结论和推测需基于我上面提供的新闻内容，并明确指出是“推测”或“分析”。\n", date))

		ctx := make([]openai.ChatCompletionMessage, 0, 4)
		ctx = append(ctx, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: content.String(),
		})
		replyMsg, err := ai_util.AskDeepSeek(ctx)
		answer := "华尔街每日资讯总结失败，请查看日志"
		if err != nil {
			answer = fmt.Sprintf("调用Deepseek失败：%s", err.Error())
		} else {
			if len(replyMsg.Choices) == 0 || len(replyMsg.Choices[0].Message.Content) == 0 {
				log.Warnf("Deepseek返回空结构：%v", replyMsg)
			} else {
				answer = replyMsg.Choices[0].Message.Content
				answer = strings.ReplaceAll(answer, "*", "")
			}
		}

		var wait sync.WaitGroup
		// 分割长消息
		messages := content_util.SplitLongMessage(answer, 1000) // 每1000字符分割一次
		for _, _group := range groups {
			wait.Add(1)
			go func(group int64) {
				defer wait.Done()
				if isGroup {
					for i, msg := range messages {
						// 添加分页标识
						pageMsg := msg
						if len(messages) > 1 {
							pageMsg = fmt.Sprintf("【%d/%d】\n%s", i+1, len(messages), msg)
						}
						bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(pageMsg)}})
						time.Sleep(1 * time.Second) // 每条消息间隔1秒
					}
				} else {
					for i, msg := range messages {
						// 添加分页标识
						pageMsg := msg
						if len(messages) > 1 {
							pageMsg = fmt.Sprintf("【%d/%d】\n%s", i+1, len(messages), msg)
						}
						bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(pageMsg)}})
						time.Sleep(1 * time.Second) // 每条消息间隔1秒
					}
				}
				time.Sleep(1 * time.Second)
			}(_group)
		}
		wait.Wait()
	}
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
