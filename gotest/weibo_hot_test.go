package gotest

import (
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/weibo_hot"
	log "github.com/sirupsen/logrus"
	"testing"
)

func Test_WeiboHot(t *testing.T) {
	html, _ := weibo_hot.GetHTML(constant.WEIBO)
	hotList, err := weibo_hot.ParseWeiboHot(html)
	if err != nil {
		log.Error(err)
	} else {
		for _, hot := range hotList {
			log.Infof("%v", hot)
		}
	}

}
