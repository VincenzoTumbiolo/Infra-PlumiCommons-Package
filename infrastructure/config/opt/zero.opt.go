package opt

// IsZero checks if val is the zero-valued
func IsZero[T comparable](val T) bool {
	var zero T
	return val == zero
}
