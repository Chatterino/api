package imgur

// NullLogger implements the go-klogger interface to ensure the go-imgur library doesn't print out a bunch of garbage
type NullLogger struct {
}

func (n NullLogger) Criticalf(format string, args ...any) {
}
func (n NullLogger) Debugf(format string, args ...any) {
}
func (n NullLogger) Errorf(format string, args ...any) {
}
func (n NullLogger) Infof(format string, args ...any) {
}
func (n NullLogger) Warningf(format string, args ...any) {
}
