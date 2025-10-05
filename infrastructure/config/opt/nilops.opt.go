package opt

import "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/ptr"

// Coalesce returns the first value if it's not nil;
// otherwise, the second
func Coalesce[T any](val *T, def T) T {
	if val != nil {
		return *val
	}

	return def
}

// MapSkipNil applies the given mapper only if the value is not nil.
// Returns [nil] otherwise
func MapSkipNil[I, O any](val *I, mapper func(I) O) *O {
	if val == nil {
		return nil
	}

	return ptr.Const(mapper(*val))
}

// FirstNonNil returns the first non-nil value it finds
// in the slice. Returns nil with no matches
func FirstNonNil[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}

	return nil
}

// AllNil checks if all the values are nil
func AllNil[T any](values ...*T) bool {
	return FirstNonNil(values...) == nil
}
