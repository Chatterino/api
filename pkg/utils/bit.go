package utils

func HasBits(sum int32, bit int32) bool {
	return (sum & bit) == bit
}
