package handler

import (
	"gapp/entity"

	"github.com/sirupsen/logrus"
)

type App struct {
	Logger *logrus.Logger
	User   entity.User
}

type HandlerResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Error   bool        `json:"error"`
	Message string      `json:"message"`
}
