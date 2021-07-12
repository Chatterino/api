package utils

// HasBits checks if sum contains bit by performing a bitwise AND operation between values
func HasBits(sum int32, bit int32) bool {
	return (sum & bit) == bit
}
