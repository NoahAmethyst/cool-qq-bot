package main

import (
	"github.com/Mrs4s/go-cqhttp/cmd/gocq"
	"time"

	_ "github.com/Mrs4s/go-cqhttp/db/leveldb"   // leveldb
	_ "github.com/Mrs4s/go-cqhttp/modules/mime" // mime检查模块
	_ "github.com/Mrs4s/go-cqhttp/modules/silk" // silk编码模块
	// 其他模块
	// _ "github.com/Mrs4s/go-cqhttp/db/mongodb"    // mongodb 数据库支持
	// _ "github.com/Mrs4s/go-cqhttp/modules/pprof" // pprof 性能分析
)

func main() {
	var cstZone = time.FixedZone("CST", 8*3600) // 东八
	time.Local = cstZone
	gracefulShutdown()
	gocq.Main()
}
