package ptr

// Assert does a double type-assertion:
// it checks if the value is of type T or *T
func Assert[T any](val any) (*T, bool) {
	valCast, ok := val.(*T)
	if !ok {
		valCastNoPtr, ok := val.(T)
		if !ok {
			return nil, false
		}

		valCast = &valCastNoPtr
	}

	return valCast, true
}

// Const is useful when you need an address of
// a constant inline, like this:
//
//	strPtr := ptr.Const("helloooo")
func Const[T any](val T) *T {
	return &val
}

// Deref the given value. Useful in mappers
func Deref[T any](val *T) T {
	return *val
}
