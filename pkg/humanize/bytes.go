package humanize

import "fmt"

var bytesSteps = []string{"B", "KB", "MB", "GB"}

func Bytes(n uint64) string {
	div := float64(1000)
	val := float64(n)
	for _, step := range bytesSteps {
		if val < 1000 {
			return fmt.Sprintf("%.1f %s", val, step)
		}
		val /= div
	}
	return fmt.Sprintf("%.1f %s", val, "TB")
}
