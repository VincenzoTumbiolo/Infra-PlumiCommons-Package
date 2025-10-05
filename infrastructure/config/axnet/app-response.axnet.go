package axnet

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/apperrors"
)

// Response is the netlayer response schema
type (
	Response[T any] struct {
		Success bool `json:"success" example:"true"`
		Result  T    `json:"result"`
	}

	// ErrorResult is the netlayer error schema
	ErrorResult struct {
		Code    string `json:"code" example:"GEN-000"`
		Message string `json:"message" example:"Unknown error"`
	}

	ErrorResponse = Response[ErrorResult]

	OkResponse struct {
		Success bool `json:"success" example:"true"`
	}
)

// GenerateResponse builds a `Response` with the given payload
func GenerateResponse[T any](result T) Response[T] {
	return Response[T]{
		Success: true,
		Result:  result,
	}
}

// GenerateOkResponse builds a `Response` with only `success = true`
func GenerateOkResponse() OkResponse {
	return OkResponse{Success: true}
}

// GenerateResponse builds a `ErrResponse` with the given payload
func GenerateErrResponse[T any](result T) Response[T] {
	return Response[T]{
		Success: false,
		Result:  result,
	}
}

// GenerateResult builds a `ErrorResult` with given error
func GenerateErrorResult(err error) ErrorResult {
	var result ErrorResult

	var appErr apperrors.AppError
	var fibErr *fiber.Error

	if errors.As(err, &appErr) {
		result = ErrorResult{
			Code:    apperrors.FullCode(appErr),
			Message: topError(err).Error(),
		}
	} else if errors.As(err, &fibErr) {
		result = ErrorResult{
			Code:    apperrors.FullCode(apperrors.GENERIC_ERROR),
			Message: fibErr.Message,
		}
	} else {
		result = ErrorResult{
			Code:    apperrors.FullCode(apperrors.GENERIC_ERROR),
			Message: apperrors.GENERIC_ERROR.Error(),
		}
	}

	switch errors.As(err, &appErr) {
	case true:
		result = ErrorResult{
			Code:    apperrors.FullCode(appErr),
			Message: topError(err).Error(),
		}

	case false:
		result = ErrorResult{
			Code:    apperrors.FullCode(apperrors.GENERIC_ERROR),
			Message: apperrors.GENERIC_ERROR.Error(),
		}
	}
	return result
}

// GenerateResponse builds a `ErrorResponse` with the given error
func GenerateErrorResponse(err error) ErrorResponse {
	result := GenerateErrorResult(err)

	return ErrorResponse{
		Success: false,
		Result:  result,
	}
}

func topError(err error) error {
	joinErr, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return err
	}

	errs := joinErr.Unwrap()
	if len(errs) == 0 {
		return err
	}

	return joinErr.Unwrap()[0]
}
