package main

import (
	"encoding/json"
	"time"
)

func mustMarshal(i interface{}) (data []byte) {
	var err error

	data, err = json.Marshal(i)
	if err != nil {
		panic(err)
	}

	return
}

func marshalNoDur(i interface{}) ([]byte, error, time.Duration) {
	data, err := json.Marshal(i)
	return data, err, noSpecialDur
}
