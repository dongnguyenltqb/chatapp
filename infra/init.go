package infra

import (
	"chatapp/util/logger"
	"context"
)

func Setup() {
	err := GetRedis().Info(context.Background(), "server").Err()
	if err != nil {
		panic(err)
	} else {
		logger.Get().Info("Redis-server: connected.")
	}

}
