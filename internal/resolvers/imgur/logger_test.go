package imgur

import (
	"testing"

	"github.com/koffeinsource/go-imgur"
)

func TestNullLogger(t *testing.T) {
	logger := &NullLogger{}

	// more of a compile-time check, ensure it fulfills the interface
	logger.Criticalf("a")
	logger.Debugf("a")
	logger.Errorf("a")
	logger.Infof("a")
	logger.Warningf("a")

	// Ensure we can instansiate an imgur client with this logger

	_ = &imgur.Client{
		Log: logger,
	}
}
