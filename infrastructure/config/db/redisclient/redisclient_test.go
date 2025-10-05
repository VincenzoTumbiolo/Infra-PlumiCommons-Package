package redisclient

import (
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestHGetErrs(t *testing.T) {
	var errs CacheErrs
	errs = append(errs, CacheErr("1"))
	errs = append(errs, CacheErr("2"))
	errs = append(errs, CacheErr("3"))

	err := error(errs)
	assert.Equal(t, errors.Is(err, errs), true)

	var asErr CacheErrs
	assert.Equal(t, errors.As(err, &asErr), true)

	assert.Equal(t, string(asErr[2]), "3")
}
