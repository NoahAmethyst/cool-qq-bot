package gotest

import (
	"github.com/Mrs4s/go-cqhttp/constant"
	"github.com/Mrs4s/go-cqhttp/util/top_list"
	log "github.com/sirupsen/logrus"
	"testing"
)

func Test_WeiboHot(t *testing.T) {
	html, _ := top_list.GetHTML(constant.WEIBO)
	hotList, err := top_list.ParseWeiboHot(html)
	if err != nil {
		log.Error(err)
	} else {
		for _, hot := range hotList {
			log.Infof("%v", hot)
		}
	}

}
