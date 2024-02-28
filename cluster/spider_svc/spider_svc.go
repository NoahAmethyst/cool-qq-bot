package spider_svc

import (
	"context"
	"github.com/Mrs4s/go-cqhttp/cluster/rpc"
	"github.com/Mrs4s/go-cqhttp/protocol/pb/spider_pb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func SvcCli() spider_pb.SpiderServiceClient {
	return spider_pb.NewSpiderServiceClient(rpc.GetConn(rpc.CliSpider))
}

func AskBingCopilot(prompt string) (*spider_pb.SpiderResp, error) {
	c := context.Background()
	cli := SvcCli()
	answer, err := cli.AskCopilot(c, &spider_pb.SpiderReq{
		Prompt: prompt,
	})
	if err != nil {
		log.Errorf("Call bing copilot failed:%s")
	}
	if len(answer.Error) > 0 {
		log.Error("Call Bing Copilot failed:%s", answer.Error)
		err = errors.New(answer.Error)
	}
	return answer, err

}

func WeiboHot() ([]*spider_pb.WeiboHot, error) {
	c := context.Background()
	cli := SvcCli()
	answer, err := cli.WeiboHot(c, &spider_pb.SpiderReq{})
	if err != nil {
		log.Errorf("Call weibo hot failed:%s", err)
	}
	return answer.WeiboHotList, err
}
