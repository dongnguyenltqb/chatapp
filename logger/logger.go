package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var once sync.Once

func Get() *logrus.Logger {
	once.Do(func() {
		logger = logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				"FieldKeyTime":  "time",
				"FieldKeyLevel": "level",
				"FieldKeyMsg":   "msg",
			},
		})
	})
	return logger
}
