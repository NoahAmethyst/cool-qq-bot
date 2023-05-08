package coolq

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/coin"
	"github.com/Mrs4s/go-cqhttp/util/file_util"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	"github.com/tristan-club/kit/log"

	"os"
	"strings"
	"sync"
	"time"
)

var SentNews SentNewsRecord

type SentNewsRecord struct {
	SentList map[int64]map[string]time.Time
	sync.RWMutex
}

func (bot *CQBot) ReportCoinPrice(group int64, isGroup bool) {

	coinContent := fmt.Sprintf("%s 币价实时信息", time.Now().Format("2006-01-02 15:04"))
	var priceContents []string
	for _, symbol := range coin.Symbols {
		coinInfo, err := coin.Get24HPriceInfo(symbol)
		if err != nil {
			log.Error().Msgf("get %s error:%s", symbol, err)
			continue
		}
		priceContents = append(priceContents, fmt.Sprintf("\n%s \n价格：%s\n24小时涨跌幅：%s%% \n最高价：%s \n最低价：%s\n",
			coinInfo.Symbol,
			strings.ReplaceAll(coinInfo.LastPrice, "000", ""),
			coinInfo.PriceChangePercent,
			strings.ReplaceAll(coinInfo.HighPrice, "000", ""),
			strings.ReplaceAll(coinInfo.LowPrice, "000", "")))
	}

	if len(priceContents) == 0 {
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("获取币价实时信息失败，请查看日志")}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("获取币价实时信息失败，请查看日志")}})
		}

	} else {
		for _, _content := range priceContents {
			coinContent += _content
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent)}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(coinContent)}})
		}

	}

}

func (bot *CQBot) ReportWeiboHot(group int64, isGroup bool) {
	hotContent := fmt.Sprintf("%s 微博实时热搜\n", time.Now().Format("2006-01-02 15:04"))
	if hotList, err := top_list.LoadWeiboHot(); err != nil {
		log.Error().Msgf("get hot list error:%s", err.Error())
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：%s", err.Error()))}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取微博热搜失败：%s", err.Error()))}})
		}

	} else {
		for _, hot := range hotList {
			hotContent += fmt.Sprintf("%d	%s\n", hot.Rank, hot.Title)
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		}

	}
}

func (bot *CQBot) Report36kr(group int64, isGroup bool) {
	hotContent := fmt.Sprintf("%s 36氪24H热榜\n", time.Now().Format("2006-01-02 15:04"))
	if hotList, err := top_list.Load36krHot(); err != nil {
		log.Error().Msgf("get hot list error:%s", err.Error())
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取36氪热榜失败：%s", err.Error()))}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
				fmt.Sprintf("爬取36氪热榜失败：%s", err.Error()))}})
		}

	} else {
		for _i, _hot := range hotList {
			if _i > 10 {
				break
			}
			hotContent += fmt.Sprintf("%d	%s\n%s\n\n", _hot.Rank, _hot.Title, _hot.Url)
		}
		if isGroup {
			bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		} else {
			bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(hotContent)}})
		}

	}
}

func (bot *CQBot) ReportWallStreetNews(group int64, isGroup bool) {
	if hotList, err := top_list.LoadWallStreetNews(); err != nil {
		log.Error().Msgf("爬取华尔街见闻最新资讯失败：%s", err.Error())
		//bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(
		//	fmt.Sprintf("爬取华尔街见闻最新资讯失败：%s", err.Error()))}})
	} else {

		var readyData []top_list.WallStreetNews
		for _, _data := range hotList {
			if !SentNews.checkSent(group, _data.Title) {
				readyData = append(readyData, _data)
				SentNews.add(group, _data.Title)
			}
		}

		if len(readyData) == 0 {
			log.Warn().Msgf("华尔街见闻：没有最新资讯，爬取资讯数量:%d", len(hotList))
			if !isGroup {
				bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText("没有华尔街最新资讯")}})
			}
		} else {
			SentNews.SaveCache()
			//倒序输出，因为最新资讯在第一个
			for i := len(readyData) - 1; i >= 0; i-- {
				content := fmt.Sprintf("%s\n\n摘要：%s\n\n链接：%s", hotList[i].Title, hotList[i].Content, hotList[i].Url)
				if isGroup {
					bot.SendGroupMessage(group, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
				} else {
					bot.SendPrivateMessage(group, 0, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
				}

				time.Sleep(3 * time.Second)
			}
		}

	}
}

func (s *SentNewsRecord) add(group int64, title string) {
	s.Lock()
	defer s.Unlock()
	now := time.Now()
	if _, ok := s.SentList[group]; !ok {
		s.SentList[group] = map[string]time.Time{
			title: now,
		}
	} else {
		s.SentList[group][title] = now
	}
	if len(s.SentList[group]) > 200 {
		for _title, _createdAt := range s.SentList[group] {
			if now.Sub(_createdAt) > 24*time.Hour {
				delete(s.SentList[group], _title)
			}
		}
	}
}

func (s *SentNewsRecord) checkSent(group int64, title string) bool {
	s.RLock()
	defer s.RUnlock()
	if v, ok := s.SentList[group]; !ok {
		return ok
	} else {
		_, ok := v[title]
		return ok
	}

}

func (s *SentNewsRecord) SaveCache() {
	s.RLock()
	defer s.RUnlock()
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}
	_, err := file_util.WriteJsonFile(s.SentList, path, "wallStreetCache", false)
	if err != nil {
		log.Error().Fields(map[string]interface{}{
			"action": "save wall street news to file",
			"error":  err,
		}).Send()
	} else {
		_ = file_util.TCCosUpload("cache", "wallStreetCache.json", fmt.Sprintf("%s/%s", path, "wallStreetCache.json"))
	}
}

func init() {
	SentNews = SentNewsRecord{
		SentList: map[int64]map[string]time.Time{},
		RWMutex:  sync.RWMutex{},
	}
	data := make(map[int64]map[string]time.Time)
	path := os.Getenv(constant.FILE_ROOT)
	if len(path) == 0 {
		path = "/tmp"
	}
	if err := file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data); err != nil {
		log.Info().Fields(map[string]interface{}{
			"action": "retry load wallstreet json from tencent cos",
		}).Send()
		_err := file_util.TCCosDownload("cache", "wallStreetCache.json", fmt.Sprintf("%s/%s", path, "wallStreetCache.json"))
		if _err == nil {
			_ = file_util.LoadJsonFile(fmt.Sprintf("%s/wallStreetCache.json", path), &data)
		}
	}
	if len(data) > 0 {
		SentNews.SentList = data
	}
}
