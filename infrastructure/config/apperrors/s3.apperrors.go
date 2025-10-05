package apperrors

import (
	"net/http"
)

type S3Error uint

const (
	S3_PRESIGNED_URL_NOT_FOUND S3Error = iota
	S3_UPLOAD_FILE
)

func (S3Error) Subject() string {
	return "S3E"
}

func (err S3Error) Code() uint {
	return uint(err)
}

func (err S3Error) StatusCode() int {
	switch err {
	case S3_PRESIGNED_URL_NOT_FOUND:
		return http.StatusNotFound
	case S3_UPLOAD_FILE:
		return http.StatusInternalServerError

	}

	return http.StatusInternalServerError
}

func (err S3Error) Error() string {
	switch err {
	case S3_PRESIGNED_URL_NOT_FOUND:
		return "Failed to get presigned URL"
	case S3_UPLOAD_FILE:
		return "Failed to upload file to S3"
	}

	return UNKNOWN_ERROR_MESSAGE
}
