package infra

import (
	"context"
)

func Setup() {
	err := GetRedis().Ping(context.Background()).Err()
	if err != nil {
		panic(err)
	}

}
