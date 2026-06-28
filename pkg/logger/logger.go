package logger

import "go.uber.org/zap"

func New(dev bool) *zap.Logger {
	if dev {
		l, _ := zap.NewDevelopment()
		return l
	}
	l, _ := zap.NewProduction()
	return l
}
