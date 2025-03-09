package logger

import (
	"go.uber.org/zap"
)

func New() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	logger.Info("logger initialized",
		zap.String("service", "packs calculator"),
	)

	return logger, nil
}
