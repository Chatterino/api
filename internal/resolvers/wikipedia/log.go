package wikipedia

import "github.com/Chatterino/api/internal/logger"

var (
	log logger.Logger
)

func SetLogger(newLog logger.Logger) {
	log = newLog
}
