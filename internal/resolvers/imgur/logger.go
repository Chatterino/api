package imgur

// NullLogger implements the go-klogger interface to ensure the go-imgur library doesn't print out a bunch of garbage
type NullLogger struct {
}

func (n NullLogger) Criticalf(format string, args ...interface{}) {
}
func (n NullLogger) Debugf(format string, args ...interface{}) {
}
func (n NullLogger) Errorf(format string, args ...interface{}) {
}
func (n NullLogger) Infof(format string, args ...interface{}) {
}
func (n NullLogger) Warningf(format string, args ...interface{}) {
}
