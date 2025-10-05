package network

import (
	"context"
	"errors"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/apperrors"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/axnet"

	"github.com/aws/aws-lambda-go/events"
)

var DEFAULT_HEADERS = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "OPTIONS,POST",
	"Access-Control-Allow-Headers": "*",
}

// APIGatewayRes builds a response compliant to the AWS APIGateway
func APIGatewayRes[T any](res T, err error) events.APIGatewayProxyResponse {
	return APIGatewayResCtx(context.Background(), res, err)
}

// APIGatewayResCtx builds a response compliant to the AWS APIGateway
func APIGatewayResCtx[T any](ctx context.Context, res T, err error) events.APIGatewayProxyResponse {
	return APIGatewayStatusResCtx(context.Background(), res, err, nil)
}

// APIGatewayStatusResCtx builds a response compliant to the AWS APIGateway
func APIGatewayStatusResCtx[T any](
	ctx context.Context,
	res T,
	err error,
	statusOpt *int,
) events.APIGatewayProxyResponse {
	if err != nil {
		body, _ := json.MarshalContext(ctx, axnet.GenerateErrorResponse(err))

		var apperr apperrors.AppError
		switch errors.As(err, &apperr) {
		case true:
			return events.APIGatewayProxyResponse{
				Headers:    DEFAULT_HEADERS,
				StatusCode: apperr.StatusCode(),
				Body:       string(body),
			}
		case false:
			return events.APIGatewayProxyResponse{
				Headers:    DEFAULT_HEADERS,
				StatusCode: http.StatusInternalServerError,
				Body:       string(body),
			}
		}
	}

	body, err := json.MarshalContext(ctx, axnet.GenerateResponse(res))
	if err != nil {
		body, _ := json.MarshalContext(ctx, axnet.GenerateErrorResponse(err))

		return events.APIGatewayProxyResponse{
			Headers:    DEFAULT_HEADERS,
			StatusCode: http.StatusInternalServerError,
			Body:       string(body),
		}
	}

	status := http.StatusOK
	if statusOpt != nil {
		status = *statusOpt
	}

	return events.APIGatewayProxyResponse{
		Headers:    DEFAULT_HEADERS,
		StatusCode: status,
		Body:       string(body),
	}
}
