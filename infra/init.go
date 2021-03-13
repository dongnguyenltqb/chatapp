package infra

import (
	"chatapp/logger"
	"context"
)

func Setup() {
	v := GetRedis().Ping(context.Background()).Val()
	logger.Get().Info(v)
}
