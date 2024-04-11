package main

import (
	"github.com/NoahAmethyst/go-cqhttp/cluster/rpc"
	"github.com/NoahAmethyst/go-cqhttp/cmd/gocq"
	"github.com/NoahAmethyst/go-cqhttp/util/cron"
	"github.com/NoahAmethyst/go-cqhttp/util/top_list"
	"github.com/sirupsen/logrus"
	"time"

	_ "github.com/NoahAmethyst/go-cqhttp/db/leveldb"   // leveldb
	_ "github.com/NoahAmethyst/go-cqhttp/modules/mime" // mime检查模块
	_ "github.com/NoahAmethyst/go-cqhttp/modules/silk" // silk编码模块
	// 其他模块
	// _ "github.com/NoahAmethyst/go-cqhttp/db/mongodb"    // mongodb 数据库支持
	// _ "github.com/NoahAmethyst/go-cqhttp/modules/pprof" // pprof 性能分析
)

func main() {
	var cstZone = time.FixedZone("CST", 8*3600) // 东八
	time.Local = cstZone
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	// Initialize grpc client
	for _, rpcCli := range rpc.RpcCliList {
		rpc.InitGrpcCli(rpcCli)
	}
	gracefulShutdown()
	cron.AddCronJob(top_list.UploadDailyRecord, "0 55 23 * * *")
	gocq.Main()
}
