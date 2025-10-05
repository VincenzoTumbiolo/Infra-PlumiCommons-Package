package utility

import (
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/vincenzotumbiolo/infra-pulumicommons-package/infrastructure/config/apperrors"
)

const JWT_ISSUER = "ax-global-auth"

type JwtConfig struct {
	Secret                       string
	TokenExpirationHours         int
	RefreshTokenExpirationMonths int
}

// GeneratePolicy builds an authorizer policy response compliant to the AWS APIGateway
func GeneratePolicy(principalID, effect, resource string, err error) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
	}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Resource: []string{resource},
					Effect:   effect,
				},
			},
		}
	}

	if err != nil {
		var apperr apperrors.AppError
		if errors.As(err, &apperr) {
			authResponse.Context = map[string]interface{}{
				"errorCode":    apperr.StatusCode(),
				"errorType":    apperr.Code(),
				"errorMessage": apperr.Error(),
			}
		}
	}

	return authResponse
}
