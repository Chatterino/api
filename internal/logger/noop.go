package logger

type NoopLogger struct {
}

func (l *NoopLogger) Debugw(msg string, keysAndValues ...interface{}) {
}

func (l *NoopLogger) Infow(msg string, keysAndValues ...interface{}) {
}

func (l *NoopLogger) Warnw(msg string, keysAndValues ...interface{}) {
}

func (l *NoopLogger) Errorw(msg string, keysAndValues ...interface{}) {
}

func (l *NoopLogger) Fatal(args ...interface{}) {
}

func (l *NoopLogger) Fatalw(msg string, keysAndValues ...interface{}) {
}

func (l *NoopLogger) Sync() error {
	return nil
}

func NewNoop() Logger {
	return &NoopLogger{}
}
