package xylog

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Logger *logrus.Logger

func InitLog(logLevel logrus.Level, logPath string) {
	Logger = logrus.New()
	Logger.SetLevel(logLevel)
	Logger.SetFormatter(&logrus.TextFormatter{})
	if len(logPath) > 0 {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			Logger.SetOutput(io.Writer(os.Stdout))
			return
		}
		Logger.SetOutput(file)
	} else {
		Logger.SetOutput(io.Writer(os.Stdout))
	}
}
