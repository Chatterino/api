package utils

import (
	"encoding/json"
	"time"

	"github.com/Chatterino/api/pkg/cache"
)

func MarshalNoDur(i interface{}) ([]byte, error, time.Duration) {
	data, err := json.Marshal(i)
	return data, err, cache.NoSpecialDur
}
