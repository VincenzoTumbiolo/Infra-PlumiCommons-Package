package apperrors

import (
	"net/http"
)

type FatalError uint

const FATAL_ERROR FatalError = 0

func (FatalError) Subject() string {
	return "FTL"
}

func (err FatalError) Code() uint {
	return uint(err)
}

func (err FatalError) StatusCode() int {
	return http.StatusInternalServerError
}

func (err FatalError) Error() string {
	return "Critical error. This instance will shut down."
}
