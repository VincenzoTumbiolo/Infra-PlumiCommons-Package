// # Optionals
//
// [Value] provide a safe way to work with values that could be empty
// without relying on pointers.
package opt

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

type Value[T any] struct {
	inner   T
	present bool
}

// New returns the provided value, wrapped in an [Value]
func New[T any](val T) Value[T] {
	return Value[T]{inner: val, present: true}
}

// Empty returns a new empty [Value]
func Empty[T any]() Value[T] {
	return Value[T]{}
}

// Wrap builds an optional from an unwrapped state.
//
// It should be used with functions that return a [bool] flag, like so:
//
//	 func GetFoo() (Foo, bool) {
//			// ...
//			return Foo{"bar"}, true
//	 }
//
//	 var foo = flow.Wrap(GetFoo())
func Wrap[T any](val T, present bool) Value[T] {
	return Value[T]{inner: val, present: present}
}

// From builds an optional from a comparable;
// if it's "zero-valued", the optional is empty
func From[T comparable](val T) Value[T] {
	if IsZero(val) {
		return Value[T]{}
	}

	return New(val)
}

// FromPtr builds an optional from a value pointer
func FromPtr[T any](valPtr *T) Value[T] {
	if valPtr == nil {
		return Value[T]{}
	}

	return New(*valPtr)
}

// IsPresent checks if the [Value] has a value
func (val Value[T]) IsPresent() bool {
	return val.present
}

// IsEmpty checks if the [Value] is empty
func (val Value[T]) IsEmpty() bool {
	return !val.IsPresent()
}

// Get unwraps the [Value]
func (val Value[T]) Get() (T, bool) {
	return val.inner, val.present
}

// Set the inner value of the [Value]
func (val *Value[T]) Set(newVal T) {
	val.inner = newVal
	val.present = true
}

// Map returns a new optional with the mapper function applied to it.
//
// Returns an empty optional if the source optional was already empty.
func Map[I, O any](opt Value[I], mapper func(I) O) Value[O] {
	if !opt.present {
		return Value[O]{}
	}

	return New(mapper(opt.inner))
}

// FlatMap returns a new optional with the mapper function applied to it.
//
// Returns an empty optional if the source optional was already empty
// or if the mapped value is also empty.
func FlatMap[I, O any](opt Value[I], mapper func(I) Value[O]) Value[O] {
	if !opt.present {
		return Value[O]{}
	}

	return mapper(opt.inner)
}

// Empty out the [Value]
func (val *Value[T]) Empty() {
	var zero T

	val.inner = zero
	val.present = false
}

// Or returns the inner value if present;
// otherwise, it returns the provided default value
func (val Value[T]) Or(def T) T {
	if val.present {
		return val.inner
	}

	return def
}

// OrZero returns the inner value if present;
// otherwise, it returns the zero value for T
func (val Value[T]) OrZero() T {
	if val.present {
		return val.inner
	}

	var zero T
	return zero
}

// Ptr converts the optional to a reference
func (val Value[T]) Ptr() *T {
	if val.present {
		return &val.inner
	}

	return nil
}

// First returns the first non-empty optional in the argument list
// or an empty optional if none is found
func First[T any](vals ...Value[T]) Value[T] {
	for _, val := range vals {
		if val.present {
			return val
		}
	}

	return Value[T]{}
}

// UnmarshalJSON overrides the default unmarshaling logic for the
// generic type parameter.
//
// It is assumed that a zero-value has to be treated as an empty [Value].
func (val *Value[T]) UnmarshalJSON(data []byte) error {
	var innerVal T
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	if err := json.Unmarshal(data, &innerVal); err != nil {
		return err
	}

	val.inner = innerVal
	val.present = true

	return nil
}

// UnmarshalJSON overrides the default unmarshaling logic for the
// generic type parameter.
//
// If a value is present, the default marshaling is applied to the inner value;
// otherwise, `null` is returned.
func (val Value[T]) MarshalJSON() ([]byte, error) {
	if val.IsEmpty() {
		return []byte("null"), nil
	}

	return json.Marshal(val.inner)
}

// IsZero is used for the encoding/json new feature `omitzero`
func (val Value[T]) IsZero() bool {
	return !val.present
}

// Scan implements the [sql.Scanner] interface
func (val *Value[T]) Scan(src any) error {
	if srcStr, ok := src.(string); ok && srcStr == "" {
		val.Empty()
		return nil
	}

	var valSql sql.Null[T]
	if err := valSql.Scan(src); err != nil {
		return err
	}

	val.inner = valSql.V
	val.present = valSql.Valid

	return nil
}

// Value implements the [driver.Valuer] interface
func (n Value[T]) Value() (driver.Value, error) {
	if !n.present {
		return nil, nil
	}

	return n.inner, nil
}
