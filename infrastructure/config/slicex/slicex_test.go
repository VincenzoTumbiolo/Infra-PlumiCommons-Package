package slicex

import (
	"reflect"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	type args[I, O any] struct {
		arr    []I
		mapper func(I) O
	}
	tests := []struct {
		name string
		args args[int, string]
		want []string
	}{
		{
			name: "int to string",
			args: args[int, string]{
				arr:    []int{1, 2, 3},
				mapper: strconv.Itoa,
			},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.arr, tt.args.mapper); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type args[T any] struct {
		arr       []T
		predicate func(T) bool
	}
	tests := []struct {
		name string
		args args[int]
		want []int
	}{
		{
			name: "numbers greater than 10",
			args: args[int]{
				arr:       []int{1, 3, 5, 10, 15, 20},
				predicate: func(n int) bool { return n > 10 },
			},
			want: []int{15, 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.args.arr, tt.args.predicate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	type args[T any] struct {
		a []T
		b []T
	}
	tests := []struct {
		name string
		args args[string]
		want []string
	}{
		{
			name: "removes 'hello'",
			args: args[string]{
				a: []string{"hey", "hello", "hi"},
				b: []string{"hello"},
			},
			want: []string{"hey", "hi"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Diff(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Diff() = %v, want %v", got, tt.want)
			}
		})
	}
}

type customErr struct {
	msg string
}
