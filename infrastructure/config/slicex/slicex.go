package slicex

import (
	"slices"
)

// Map returns a new slice with every element transformed with the given mapper function
func Map[S ~[]I, I, O any](arr S, mapper func(I) O) []O {
	result := make([]O, 0, len(arr))
	for _, v := range arr {
		result = append(result, mapper(v))
	}

	return result
}

// FlatMap returns a new slice with every element transformed with the given mapper function.
// The resulting slices are concatenated
func FlatMap[S ~[]I, I, O any](arr S, mapper func(I) []O) []O {
	result := make([]O, 0, len(arr))
	for _, v := range arr {
		result = append(result, mapper(v)...)
	}

	return result
}

// MapI returns a new slice with every element transformed with the given mapper function.
// The mapper function has access to the current position of the element
func MapI[S ~[]I, I, O any](arr S, mapper func(int, I) O) []O {
	result := make([]O, 0, len(arr))
	for i, v := range arr {
		result = append(result, mapper(i, v))
	}

	return result
}

// FlatMapI returns a new slice with every element transformed with the given mapper function.
// The mapper function has access to the current position of the element.
// The resulting slices are concatenated
func FlatMapI[S ~[]I, I, O any](arr S, mapper func(int, I) []O) []O {
	result := make([]O, 0, len(arr))
	for i, v := range arr {
		result = append(result, mapper(i, v)...)
	}

	return result
}

// MapErr returns a new slice with every element transformed with the given mapper function
// and a possible error
func MapErr[S ~[]I, I, O any](arr S, mapper func(I) (O, error)) ([]O, error) {
	result := make([]O, 0, len(arr))
	for _, v := range arr {
		res, err := mapper(v)
		if err != nil {
			return nil, err
		}

		result = append(result, res)
	}

	return result, nil
}

// FlatMapErr returns a new slice with every element transformed with the given mapper function
// and a possible error.
// The resulting slices are concatenated
func FlatMapErr[S ~[]I, I, O any](arr S, mapper func(I) ([]O, error)) ([]O, error) {
	result := make([]O, 0, len(arr))
	for _, v := range arr {
		res, err := mapper(v)
		if err != nil {
			return nil, err
		}

		result = append(result, res...)
	}

	return result, nil
}

// MapErrI returns a new slice with every element transformed with the given mapper function
// and a possible error.
// The mapper function has access to the current position of the element
func MapErrI[S ~[]I, I, O any](arr S, mapper func(int, I) (O, error)) ([]O, error) {
	result := make([]O, 0, len(arr))
	for i, v := range arr {
		res, err := mapper(i, v)
		if err != nil {
			return nil, err
		}

		result = append(result, res)
	}

	return result, nil
}

// FlatMapErrI returns a new slice with every element transformed with the given mapper function
// and a possible error.
// The mapper function has access to the current position of the element.
// The resulting slices are concatenated
func FlatMapErrI[S ~[]I, I, O any](arr S, mapper func(int, I) ([]O, error)) ([]O, error) {
	result := make([]O, 0, len(arr))
	for i, v := range arr {
		res, err := mapper(i, v)
		if err != nil {
			return nil, err
		}

		result = append(result, res...)
	}

	return result, nil
}

// Filter returns a new slice with only the elements that validate the predicate
func Filter[S ~[]T, T any](arr S, predicate func(T) bool) []T {
	result := []T{}
	for _, v := range arr {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

// Diff returns all the elements that are not in common between the two slices
func Diff[S ~[]T, T comparable](a S, b S) S {
	var res S
	for _, v := range a {
		if !slices.Contains(b, v) {
			res = append(res, v)
		}
	}

	return res
}

// FindFirst returns the poistion and pointer of the first element that satisfies the predicate.
// Returns -1 and nil otherwise
func FindFirst[S ~[]T, T any](arr S, predicate func(T) bool) (*T, int) {
	for i, v := range arr {
		if predicate(v) {
			return &v, i
		}
	}

	return nil, -1
}

// ContainsAll returns true if all the elements in arrContained are in arr
func ContainsAll[S ~[]T, T comparable](arr S, arrContained []T) bool {
	for _, v := range arrContained {
		if !slices.Contains(arr, v) {
			return false
		}
	}

	return true
}

// Group the slice with the given identity
func Group[S ~[]V, K comparable, V any](arr S, identity func(V) K) map[K]S {
	res := make(map[K]S, len(arr))
	for _, val := range arr {
		id := identity(val)
		res[id] = append(res[id], val)
	}

	return res
}
