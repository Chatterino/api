package utils

import (
	"encoding/json"
	"time"

	"github.com/Chatterino/api/pkg/cache"
)

func MarshalNoDur(i interface{}) ([]byte, *int, *string, time.Duration, error) {
	data, err := json.Marshal(i)
	return data, nil, nil, cache.NoSpecialDur, err
}
