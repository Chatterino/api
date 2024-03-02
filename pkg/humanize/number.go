package humanize

import (
	"bytes"
	"fmt"
	"strconv"
)

func insertCommas(str string, n int) string {
	var buffer bytes.Buffer
	var remainder = n - 1
	var lenght = len(str) - 2
	for i, rune := range str {
		buffer.WriteRune(rune)
		if (lenght-i)%n == remainder {
			buffer.WriteRune(',')
		}
	}
	return buffer.String()
}

func Number(number uint64) string {
	if number < 1_000_000 {
		return insertCommas(strconv.FormatUint(number, 10), 3)
	}

	inMillions := float64(number) / 1_000_000
	return fmt.Sprintf("%.1fM", inMillions)
}

func NumberInt64(number int64) string {
	if number < 1_000_000 {
		return insertCommas(strconv.FormatInt(number, 10), 3)
	}

	inMillions := float64(number) / 1_000_000
	return fmt.Sprintf("%.1fM", inMillions)
}
