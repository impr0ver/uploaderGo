package logger

import (
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	//SugaredLogger registrator
	sugar := *logger.Sugar()
	return &sugar
}
