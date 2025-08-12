package logger

// MockLogger is a logger that does nothing. For testing.
type MockLogger struct{}

// NewMockLogger creates a new instance of the mock logger. Doesn't fail. Ever.
func NewMockLogger() Logger {
	return &MockLogger{}
}

// Ensure MockLogger satisfies the Logger interface at compile time.
var _ Logger = (*MockLogger)(nil)

func (l *MockLogger) Debugw(msg string, keysAndValues ...any) {}
func (l *MockLogger) Infow(msg string, keysAndValues ...any)  {}
func (l *MockLogger) Warnw(msg string, keysAndValues ...any)  {}
func (l *MockLogger) Errorw(msg string, keysAndValues ...any) {}
func (l *MockLogger) Panicw(msg string, keysAndValues ...any) {}
func (l *MockLogger) Fatalw(msg string, keysAndValues ...any) {}
