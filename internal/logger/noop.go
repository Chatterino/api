package logger

type NoopLogger struct {
}

func (l *NoopLogger) Debugw(msg string, keysAndValues ...any) {
}

func (l *NoopLogger) Infow(msg string, keysAndValues ...any) {
}

func (l *NoopLogger) Warnw(msg string, keysAndValues ...any) {
}

func (l *NoopLogger) Errorw(msg string, keysAndValues ...any) {
}

func (l *NoopLogger) Fatal(args ...any) {
}

func (l *NoopLogger) Fatalw(msg string, keysAndValues ...any) {
}

func (l *NoopLogger) Sync() error {
	return nil
}

func NewNoop() Logger {
	return &NoopLogger{}
}
