package spider_svc

import (
	"context"
	"github.com/Mrs4s/go-cqhttp/cluster/rpc"
	"github.com/Mrs4s/go-cqhttp/protocol/pb/spider_pb"
	"github.com/rs/zerolog/log"
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
		log.Err(err)
	}
	return answer, err

}
