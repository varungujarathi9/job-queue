package utils

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func InitLogger() {
	logFile, err := os.OpenFile("job-queue.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		Logger.Out = logFile
	} else {
		fmt.Println("Failed to log to file, using default stderr")
	}
	Logger.SetFormatter(&logrus.JSONFormatter{})
	Logger.SetLevel(logrus.ErrorLevel)
}
