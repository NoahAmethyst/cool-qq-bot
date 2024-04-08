package spider_svc

import (
	"context"
	"github.com/NoahAmethyst/go-cqhttp/cluster/rpc"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/spider_pb"
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
	resp, err := cli.WeiboHot(c, &spider_pb.SpiderReq{})
	if err != nil {
		log.Errorf("Call weibo hot failed:%s", err)
	}
	return resp.WeiboHotList, err
}

func ZhihuHot() ([]*spider_pb.ZhihuHot, error) {
	c := context.Background()
	cli := SvcCli()
	resp, err := cli.ZhihuHot(c, &spider_pb.SpiderReq{})
	if err != nil {
		log.Errorf("Call ZhihuHot faield:%s", err.Error())
	}
	return resp.ZhihuHotList, err
}

func WallStreetNews() ([]*spider_pb.WallStreetNew, error) {
	c := context.Background()
	cli := SvcCli()
	resp, err := cli.WallStreetNews(c, &spider_pb.SpiderReq{})
	if err != nil {
		log.Errorf("Call wall street news failed:%s", err.Error())
	}
	return resp.WallStreetNews, err
}

func D36Kr() ([]*spider_pb.D36KrHot, error) {
	c := context.Background()
	cli := SvcCli()
	resp, err := cli.D36KrHot(c, &spider_pb.SpiderReq{})
	if err != nil {
		log.Errorf("Call 36KR failed:%s", err.Error())
	}
	return resp.D36KrHotList, err
}

func CaiXinNews() ([]*spider_pb.CaiXinNew, error) {
	c := context.Background()
	cli := SvcCli()
	resp, err := cli.CaiXinNews(c, &spider_pb.SpiderReq{})
	if err != nil {
		log.Errorf("Call CaiXin News failed:%s", err.Error())
	}
	return resp.CaiXinNews, err
}
