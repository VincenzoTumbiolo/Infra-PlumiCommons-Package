package apperrors

import (
	"net/http"
)

type GenericError uint

const GENERIC_ERROR GenericError = 0

func (GenericError) Subject() string {
	return "GEN"
}

func (err GenericError) Code() uint {
	return uint(err)
}

func (err GenericError) StatusCode() int {
	return http.StatusInternalServerError
}

func (err GenericError) Error() string {
	return UNKNOWN_ERROR_MESSAGE
}
