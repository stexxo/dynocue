package logging

type NoopLogger struct{}

func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

func (n NoopLogger) Debug(msg string, args ...any) {}
func (n NoopLogger) Info(msg string, args ...any)  {}
func (n NoopLogger) Warn(msg string, args ...any)  {}
func (n NoopLogger) Error(msg string, args ...any) {}
