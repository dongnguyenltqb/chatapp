package logger

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var onceInitLogger sync.Once

func Get() *logrus.Logger {
	onceInitLogger.Do(func() {
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
