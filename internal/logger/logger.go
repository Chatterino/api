package logger

// Logger defines the smallest interface we use of go-uber/zap's sugared logger
// Defining this interface makes it easier for us to control what functions we use, and to pass the logger around to various packages
type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})

	Infow(msg string, keysAndValues ...interface{})

	Warnw(msg string, keysAndValues ...interface{})

	Errorw(msg string, keysAndValues ...interface{})

	Fatal(args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}
