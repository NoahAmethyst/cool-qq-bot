package task

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Mrs4s/go-cqhttp/coolq"
	"github.com/Mrs4s/go-cqhttp/util/weibo_hot"
	log "github.com/sirupsen/logrus"
	"time"
)

func HourlyWeiboHot(bot *coolq.CQBot) {
	go func(bot *coolq.CQBot) {
		for {
			var CNZone = time.FixedZone("CST", 8*3600)
			content := fmt.Sprintf("%s 微博实时热搜\n", time.Now().In(CNZone).Format("2006-01-02 15:04:05"))

			if hotList, err := weibo_hot.Summary(); err != nil {
				log.Errorf("get hot list error:%s", err.Error())
			} else {
				for _, hot := range hotList {
					content += fmt.Sprintf("%d	%s\n", hot.Rank, hot.Title)
				}
			}
			bot.SendGroupMessage(124372891, &message.SendingMessage{Elements: []message.IMessageElement{message.NewText(content)}})
			time.Sleep(1 * time.Hour)
		}

	}(bot)
}
