package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type sugaredLogger struct {
	logger *zap.SugaredLogger
}

func NewSugarLogger() (Logger, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to create a logger: %v", err)
	}
	return &sugaredLogger{logger: l.Sugar()}, nil
}

func (l *sugaredLogger) Debugw(msg string, keysAndValues ...any) {
	l.logger.Debugw(msg, keysAndValues...)
}

func (l *sugaredLogger) Infow(msg string, keysAndValues ...any) {
	l.logger.Infow(msg, keysAndValues...)
}

func (l *sugaredLogger) Warnw(msg string, keysAndValues ...any) {
	l.logger.Warnw(msg, keysAndValues...)
}

func (l *sugaredLogger) Errorw(msg string, keysAndValues ...any) {
	l.logger.Errorw(msg, keysAndValues...)
}

func (l *sugaredLogger) Panicw(msg string, keysAndValues ...any) {
	l.logger.Panicw(msg, keysAndValues...)
}

func (l *sugaredLogger) Fatalw(msg string, keysAndValues ...any) {
	l.logger.Fatalw(msg, keysAndValues...)
}

func (l *sugaredLogger) Sync() error {
	return l.logger.Sync()
}
