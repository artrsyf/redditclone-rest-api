package tools

import "github.com/sirupsen/logrus"

var Logger *logrus.Logger

func Init() {
	Logger = logrus.New()
	Logger.SetLevel(logrus.DebugLevel)
}
