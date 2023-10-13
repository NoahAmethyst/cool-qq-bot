package middleware

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"strconv"
	"strings"
	"time"
)

func LoggerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	var requestId string
	if _uuid, err := uuid.NewUUID(); err != nil {
		requestId = strconv.FormatInt(time.Now().UnixMilli(), 10)
	} else {
		requestId = strings.ReplaceAll(_uuid.String(), "-", "")
	}

	log.Infof("[%s] Receive grpc request: Method [%s] Body [%+v]", requestId, info.FullMethod, info.Server)
	resp, err := handler(ctx, req)
	if err != nil {
		log.Warnf("[%s] Grpc response Method [%s]  failed: %s", requestId, info.FullMethod, err.Error())
	} else {
		log.Infof("[%s] Grpc response Method [%s] success:", requestId, info.FullMethod)
	}
	return resp, err
}
