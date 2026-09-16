package logger

import (
	"go.uber.org/zap"
)

func NewZapLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}