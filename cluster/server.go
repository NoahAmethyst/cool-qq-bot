package cluster

import (
	"context"
	"fmt"

	"github.com/NoahAmethyst/go-cqhttp/bot_service"
	"github.com/NoahAmethyst/go-cqhttp/cluster/middleware"
	"github.com/NoahAmethyst/go-cqhttp/constant"
	"github.com/NoahAmethyst/go-cqhttp/protocol/pb/qqbot_pb"
	log "github.com/sirupsen/logrus"

	"net"
	"runtime/debug"
	"strconv"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	customFunc grpc_recovery.RecoveryHandlerFuncContext
)

var KubeOptServer *grpc.Server

func StartServer(grpcPort string) {
	if len(grpcPort) == 0 {
		grpcPort = strconv.Itoa(constant.DefaultGRPCPort)
	}

	grpcAddr := fmt.Sprintf("0.0.0.0:%s", grpcPort)
	lis, err := net.Listen("tcp", grpcAddr)

	if err != nil {
		log.Errorf("Start grpc listenr failed:%s", err.Error())
	}
	log.Infof("QQBot service start at address %s", grpcAddr)

	// Define customfunc to handle panic
	customFunc = func(ctx context.Context, p interface{}) error {
		log.Errorf("[PANIC] %s\n\n%s", p, string(debug.Stack()))
		return status.Errorf(codes.Unknown, "panic triggered: %v", p)
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandlerContext(customFunc),
	}

	// Create a server. Recovery handlers should typically be last in the chain_info so that other middleware
	// (e.g. logging) can operate on the recovered state instead of being directly affected by any panic

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.LoggerInterceptor),
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(opts...),
			//otgrpc.OpenTracingServerInterceptor(thisTracer),
		),
		grpc.ChainStreamInterceptor(
			grpc_recovery.StreamServerInterceptor(opts...),
			//grpc_opentracing.StreamServerInterceptor(topts...),
		),
	)

	//register bot grpc server
	qqbot_pb.RegisterQQBotServiceServer(grpcServer, bot_service.BotService{})
	KubeOptServer = grpcServer

	reflection.Register(grpcServer)

	err = grpcServer.Serve(lis)

	if err != nil {
		log.Error(err)
	}
}
