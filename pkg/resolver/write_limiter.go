package resolver

import (
	"errors"
)

// WriteLimiter can limit how many bytes can be written before erroring out
type WriteLimiter struct {
	Limit uint64

	total uint64
}

func (wc *WriteLimiter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	if wc.total > wc.Limit {
		return n, errors.New("too big2")
	}
	return n, nil
}
