package apperrors

import (
	"net/http"
)

type AuthError uint

const (
	WRONG_CREDENTIALS AuthError = iota
	WRONG_ROLE_FOR_THE_OPERATION
	CAS_FORBIDDEN_FOR_THIS_USER
	TOKEN_EXPIRED
	MISSING_TOKEN
	TOKEN_MALFORMED
	TOKEN_NOT_FOUND_IN_DB
	WRONG_PRIVILEGES
	FILTERS_DO_NOT_INTERSECT_WITH_PRIVILEGES
	USER_NOT_REGISTERED_IN_SERVICE
)

func (AuthError) Subject() string {
	return "AUT"
}

func (err AuthError) Code() uint {
	return uint(err)
}

func (err AuthError) StatusCode() int {
	switch err {
	case WRONG_CREDENTIALS:
		return http.StatusUnauthorized
	case WRONG_ROLE_FOR_THE_OPERATION:
		return http.StatusForbidden
	case CAS_FORBIDDEN_FOR_THIS_USER:
		return http.StatusForbidden
	case TOKEN_EXPIRED:
		return http.StatusForbidden
	case MISSING_TOKEN:
		return http.StatusUnauthorized
	case TOKEN_MALFORMED:
		return http.StatusUnauthorized
	case TOKEN_NOT_FOUND_IN_DB:
		return http.StatusUnauthorized
	case WRONG_PRIVILEGES:
		return http.StatusNotAcceptable
	case FILTERS_DO_NOT_INTERSECT_WITH_PRIVILEGES:
		return http.StatusForbidden
	case USER_NOT_REGISTERED_IN_SERVICE:
		return http.StatusForbidden
	}

	return http.StatusUnauthorized
}

func (err AuthError) Error() string {
	switch err {
	case WRONG_CREDENTIALS:
		return "Wrong credentials"
	case WRONG_ROLE_FOR_THE_OPERATION:
		return "Wrong role for the operation"
	case CAS_FORBIDDEN_FOR_THIS_USER:
		return "Cas forbidden for this user"
	case TOKEN_EXPIRED:
		return "Token expired"
	case MISSING_TOKEN:
		return "Missing token"
	case TOKEN_MALFORMED:
		return "Token malformed"
	case TOKEN_NOT_FOUND_IN_DB:
		return "Token not found in db"
	case WRONG_PRIVILEGES:
		return "Wrong privileges"
	case FILTERS_DO_NOT_INTERSECT_WITH_PRIVILEGES:
		return "Filters do not intersect with privileges"
	case USER_NOT_REGISTERED_IN_SERVICE:
		return "User is not allowed to interact with this service"
	}

	return UNKNOWN_ERROR_MESSAGE
}
