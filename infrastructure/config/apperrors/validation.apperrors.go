package apperrors

import (
	"net/http"
)

type ValidationError uint

const (
	INVALID_JSON_BODY ValidationError = iota
	INVALID_PARAMS
	INVALID_BODY
	INVALID_QUERY
	INVALID_PARAMS_UUID
	INVALID_HEADERS
)

func (ValidationError) Subject() string {
	return "VAL"
}

func (err ValidationError) Code() uint {
	return uint(err)
}

func (err ValidationError) StatusCode() int {
	return http.StatusBadRequest
}

func (err ValidationError) Error() string {
	switch err {
	case INVALID_JSON_BODY:
		return "Body is not a valid JSON"
	case INVALID_PARAMS:
		return "Params validation error"
	case INVALID_BODY:
		return "Body validation error"
	case INVALID_QUERY:
		return "Query validation error"
	case INVALID_PARAMS_UUID:
		return "Params is not a valid UUID"
	case INVALID_HEADERS:
		return "Headers don't match the expected format"
	}

	return UNKNOWN_ERROR_MESSAGE
}
