// See https://app.clickup.com/20558920/v/dc/kkd28-45712

package assert

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"testing"
)

var AssertErr = errors.New("assertion failed")

// CrashStrategy indicates how the program should behave
// when encountering a failed assertion
type CrashStrategy int

const (
	// Nuke the program
	NukeCrashStrategy CrashStrategy = iota

	// Propagate the error via panicking. Useful for servers or testing environments
	GracefulCrashStrategy

	// Propagate the error via the Go testing framework
	TestingCrashStrategy
)

var crashStrategy = NukeCrashStrategy

func Graceful() {
	crashStrategy = GracefulCrashStrategy
}

var test testing.TB

func Testing(t testing.TB) {
	crashStrategy = TestingCrashStrategy
	test = t
}

// Must asserts that the error is nil
func Must(err error) {
	if err != nil {
		crash(fmt.Errorf("%w: error is not nil (%w)", AssertErr, err))
	}
}

// MustRes asserts that the error is nil, and returns the response
func MustRes[T any](res T, err error) T {
	if err != nil {
		crash(fmt.Errorf("%w: error is not nil (%w)", AssertErr, err))
	}

	return res
}

// MustOk asserts that the ok flag is true, and returns the response
func MustOk[T any](res T, ok bool) T {
	if !ok {
		crash(fmt.Errorf("%w: boolean flag is false", AssertErr))
	}

	return res
}

// Check asserts that the condition is true
func Check(cond bool) {
	if !cond {
		crash(fmt.Errorf("%w: condition is false", AssertErr))
	}
}

// NotZero asserts that the value is not zeroed
func NotZero[T comparable](val T) {
	var zero T
	if val == zero {
		crash(fmt.Errorf("%w: value %T is zeroed", AssertErr, val))
	}
}

// NotEmpty asserts that the slice is not nil, nor empty
func NotEmpty[T any](slice []T) {
	if len(slice) == 0 {
		crash(fmt.Errorf("%w: slice %T is empty", AssertErr, slice))
	}
}

func crash(err error) {
	_, file, line, _ := runtime.Caller(2)
	log.Printf("%s\n%s:%d\n\n", err.Error(), file, line)

	stackBuf := make([]byte, 1024*4)
	runtime.Stack(stackBuf, true)

	switch crashStrategy {
	case NukeCrashStrategy:
		log.Fatal(string(stackBuf))

	case GracefulCrashStrategy:
		log.Print(string(stackBuf))
		panic(err)

	case TestingCrashStrategy:
		test.Log(string(stackBuf))
		test.Fatal(err)
	}
}
