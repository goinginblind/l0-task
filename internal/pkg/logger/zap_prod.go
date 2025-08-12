package logger

import (
	"go.uber.org/zap"
)

type developmentLogger struct {
	logger *zap.Logger
}

// NewDevelopmentLogger creates a new instance of the development zap logger.
func NewDevelopmentLogger() (Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &developmentLogger{logger: zapLogger}, nil
}

func (l *developmentLogger) toZapFields(keysAndValues ...any) []zap.Field {
	var fields []zap.Field
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			fields = append(fields, zap.Any(keysAndValues[i].(string), keysAndValues[i+1]))
		}
	}
	return fields
}

func (l *developmentLogger) Debugw(msg string, keysAndValues ...any) {
	l.logger.Debug(msg, l.toZapFields(keysAndValues...)...)
}

func (l *developmentLogger) Infow(msg string, keysAndValues ...any) {
	l.logger.Info(msg, l.toZapFields(keysAndValues...)...)
}

func (l *developmentLogger) Warnw(msg string, keysAndValues ...any) {
	l.logger.Warn(msg, l.toZapFields(keysAndValues...)...)
}

func (l *developmentLogger) Errorw(msg string, keysAndValues ...any) {
	l.logger.Error(msg, l.toZapFields(keysAndValues...)...)
}

func (l *developmentLogger) Panicw(msg string, keysAndValues ...any) {
	l.logger.Panic(msg, l.toZapFields(keysAndValues...)...)
}

func (l *developmentLogger) Fatalw(msg string, keysAndValues ...any) {
	l.logger.Fatal(msg, l.toZapFields(keysAndValues...)...)
}

func (l *developmentLogger) Sync() error {
	return l.logger.Sync()
}
