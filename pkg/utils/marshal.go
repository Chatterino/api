package utils

import (
	"encoding/json"
	"time"

	"github.com/Chatterino/api/pkg/cache"
)

func MarshalNoDur(i interface{}) ([]byte, time.Duration, error) {
	data, err := json.Marshal(i)
	return data, cache.NoSpecialDur, err
}
