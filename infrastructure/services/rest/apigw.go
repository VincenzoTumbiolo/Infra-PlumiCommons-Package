package rest

import (
	"fmt"

	policy "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/aws"
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"log/slog"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
)

func (mod RESTModule) CreateRestAPI(input dto.CreateRestAPIInput) error {
	restApi, err := apigateway.NewRestApi(mod.Ctx, fmt.Sprintf("%s-rest-api", input.BaseName), &apigateway.RestApiArgs{
		Name: pulumi.StringPtr(fmt.Sprintf("%s-rest-api", input.BaseName)),
		Tags: input.Tags,
	})
	if err != nil {
		slog.Error("Failed to Create Rest API", "err: ", err)
		return err
	}
	var authorizer *apigateway.Authorizer
	var authID pulumi.StringPtrInput = nil
	if input.LambdaAuth != nil {
		authorizer, err = mod.CreateAuthorizer(dto.CreateAuthorizerInput{
			CreateRestAPI: dto.CreateRestAPI{

				BaseName:  input.BaseName,
				Region:    input.Region,
				AccountId: input.AccountId,
			},
			RestApi:        restApi,
			Tags:           input.Tags,
			LambdaAuth:     input.LambdaAuth,
			IdentitySource: input.IdentitySource,
		})
		if err != nil {
			slog.Error("Failed to Create API Authorizer", "err: ", err)
			return err
		}
		authID = authorizer.ID().ToStringPtrOutput()
	}

	allOptions := make([]pulumi.Resource, 0)
	if len(input.Endpoints) != 0 {
		allOptions, err = mod.CreateEndpoints(dto.CreateEndpointsInput{
			RestApi:    restApi,
			Endpoints:  input.Endpoints,
			AllOptions: allOptions,
			AuthID:     &authID,
			CreateRestAPI: dto.CreateRestAPI{
				BaseName:  input.BaseName,
				Region:    input.Region,
				AccountId: input.AccountId,
			},
		},
		)
		if err != nil {
			slog.Error("Failed to Create API Endpoints", "err: ", err)
			return err
		}
	}

	deploy, err := apigateway.NewDeployment(mod.Ctx, fmt.Sprintf("%s-deploy", input.BaseName), &apigateway.DeploymentArgs{
		RestApi: restApi.ID(),
	}, pulumi.DependsOn(allOptions))
	if err != nil {
		slog.Error("Failed to Create API Deploy", "err: ", err)
		return err
	}

	allOptions = append(allOptions, deploy)

	domain, err := apigateway.NewDomainName(mod.Ctx, fmt.Sprintf("%s-domain", input.BaseName), &apigateway.DomainNameArgs{
		DomainName:             pulumi.String(input.DomainName),
		RegionalCertificateArn: pulumi.String(fmt.Sprintf("arn:aws:acm:%s:%s:certificate/%s", input.Region, input.AccountId, input.HttpsCertificateId)),
		EndpointConfiguration: &apigateway.DomainNameEndpointConfigurationArgs{
			Types: pulumi.String("REGIONAL"),
		},
		Tags: input.Tags,
	})
	if err != nil {
		slog.Error("Failed to Create API Custom Domain Name", "err: ", err)
		return err
	}

	stage, err := apigateway.NewStage(mod.Ctx, fmt.Sprintf("%s-stage", input.BaseName), &apigateway.StageArgs{
		RestApi:     restApi.ID(),
		Deployment:  deploy.ID(),
		StageName:   pulumi.String(input.StageName),
		Description: pulumi.String(input.BaseName),
	}, pulumi.DependsOn(allOptions))
	if err != nil {
		slog.Error("Failed to Create API Stage", "err: ", err)
		return err
	}

	assumeRole, err := iam.GetPolicyDocument(mod.Ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: pulumi.StringRef("Allow"),
				Principals: []iam.GetPolicyDocumentStatementPrincipal{
					{
						Type: "Service",
						Identifiers: []string{
							"apigateway.amazonaws.com",
						},
					},
				},
				Actions: []string{
					"sts:AssumeRole",
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}
	cloudwatchRole, err := iam.NewRole(mod.Ctx, fmt.Sprintf("%s-apigw-cloudwatch-global", input.BaseName), &iam.RoleArgs{
		Name:             pulumi.String(fmt.Sprintf("%s-apigw-cloudwatch-global", input.BaseName)),
		AssumeRolePolicy: pulumi.String(assumeRole.Json),
	})
	if err != nil {
		return err
	}
	_, err = apigateway.NewAccount(mod.Ctx, fmt.Sprintf("%s-apigw-account", input.BaseName), &apigateway.AccountArgs{
		CloudwatchRoleArn: cloudwatchRole.Arn,
	})
	if err != nil {
		return err
	}
	cloudwatch, err := iam.GetPolicyDocument(mod.Ctx, &iam.GetPolicyDocumentArgs{
		Statements: []iam.GetPolicyDocumentStatement{
			{
				Effect: pulumi.StringRef("Allow"),
				Actions: []string{
					"logs:CreateLogGroup",
					"logs:CreateLogStream",
					"logs:DescribeLogGroups",
					"logs:DescribeLogStreams",
					"logs:PutLogEvents",
					"logs:GetLogEvents",
					"logs:FilterLogEvents",
				},
				Resources: []string{
					"*",
				},
			},
		},
	}, nil)
	if err != nil {
		return err
	}
	_, err = iam.NewRolePolicy(mod.Ctx, "cloudwatch", &iam.RolePolicyArgs{
		Name:   pulumi.String("default"),
		Role:   cloudwatchRole.ID(),
		Policy: pulumi.String(cloudwatch.Json),
	})
	if err != nil {
		return err
	}
	_, err = apigateway.NewAccount(mod.Ctx, "account", &apigateway.AccountArgs{
		CloudwatchRoleArn: cloudwatchRole.Arn,
	})
	if err != nil {
		slog.Error("Failed to Create Api GW Account for resource", "err: ", err)
		return err
	}

	_, err = apigateway.NewMethodSettings(mod.Ctx, "methodSettings", &apigateway.MethodSettingsArgs{
		RestApi:   restApi.ID(),
		StageName: stage.StageName,
		Settings: apigateway.MethodSettingsSettingsArgs{
			LoggingLevel:     pulumi.String("INFO"),
			MetricsEnabled:   pulumi.Bool(true),
			DataTraceEnabled: pulumi.Bool(true),
		},
		MethodPath: pulumi.String("*/*"),
	})
	if err != nil {
		slog.Error("Failed to Create API Stage Method settings", "err", err)
		return err
	}

	_, err = apigateway.NewBasePathMapping(mod.Ctx, fmt.Sprintf("%s-path-mapping", input.BaseName), &apigateway.BasePathMappingArgs{
		RestApi:    restApi.ID(),
		StageName:  stage.StageName,
		DomainName: domain.DomainName,
	})
	if err != nil {
		slog.Error("Failed to Create API Path Mapping", "err: ", err)
		return err
	}

	return nil
}

func (mod RESTModule) CreateEndpoints(input dto.CreateEndpointsInput) ([]pulumi.Resource, error) {
	if err := validateUniqueEndpoints(input.Endpoints); err != nil {
		slog.Error("Duplicate endpoints detected", "err", err)
		return nil, err
	}

	for i := 0; i < len(input.Endpoints); i++ {
		endpoint := input.Endpoints[i]
		baseResource, err := apigateway.NewResource(mod.Ctx, endpoint.Name, &apigateway.ResourceArgs{
			RestApi:  input.RestApi.ID(),
			ParentId: input.RestApi.RootResourceId,
			PathPart: pulumi.String(endpoint.Path),
		})
		if err != nil {
			slog.Error("Failed to Create Resource Rest API", "err: ", err)
			return nil, err
		}
		integrationResponse, optionsMethod, err := mod.CreateOptions(dto.CreateOptionsInput{
			RestApi:      input.RestApi,
			BaseResource: baseResource,
			Endpoint:     endpoint,
		})
		if err != nil {
			return nil, err
		}

		input.AllOptions = append(input.AllOptions, integrationResponse, optionsMethod)
		for i := 0; i < len(endpoint.Methods); i++ {
			method := endpoint.Methods[i]
			err = mod.CreateMethodIntegration(dto.CreateMethodIntegrationInput{
				ApiID:           input.RestApi.ID(),
				Name:            method.Name,
				RootResourceID:  baseResource.ID(),
				HttpMethod:      method.HttpMethod,
				TargetLambdaArn: method.TargetLambdaInvokeArn,
				AuthorizerId:    input.AuthID,
			})
			if err != nil {
				slog.Error("Failed to Create "+method.Name+" Method Rest API", "err: ", err)
				return nil, err
			}
			_, err := lambda.NewPermission(mod.Ctx, fmt.Sprintf("%s-%s-lambda-permission", input.BaseName, method.Name), &lambda.PermissionArgs{
				Action:    pulumi.String("lambda:InvokeFunction"),
				Function:  method.TargetLambdaFunctionName,
				Principal: pulumi.String("apigateway.amazonaws.com"),
				SourceArn: pulumi.Sprintf("arn:aws:execute-api:%s:%s:%s/*/%s%s", input.Region, input.AccountId, input.RestApi.ID(), method.HttpMethod, baseResource.Path),
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return input.AllOptions, nil
}

func (mod RESTModule) CreateOptions(input dto.CreateOptionsInput) (*apigateway.IntegrationResponse, *apigateway.Method, error) {
	optionsMethod, err := apigateway.NewMethod(mod.Ctx, fmt.Sprintf("%s-optionsMethod", input.Endpoint.Name), &apigateway.MethodArgs{
		RestApi:       input.RestApi.ID(),
		ResourceId:    input.BaseResource.ID(),
		HttpMethod:    pulumi.String("OPTIONS"),
		Authorization: pulumi.String("NONE"),
		AuthorizerId:  nil,
	})
	if err != nil {
		slog.Error("Failed to Create OPTIONS", "err: ", err)
		return nil, nil, err
	}

	integration, err := apigateway.NewIntegration(mod.Ctx, fmt.Sprintf("%s-optionsIntegration", input.Endpoint.Name), &apigateway.IntegrationArgs{
		RestApi:               input.RestApi.ID(),
		ResourceId:            input.BaseResource.ID(),
		HttpMethod:            pulumi.String("OPTIONS"),
		IntegrationHttpMethod: pulumi.String("POST"),
		Type:                  pulumi.String("MOCK"),
		RequestTemplates: pulumi.StringMap{
			"application/json": pulumi.String("{\"statusCode\": 200}"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{optionsMethod}))
	if err != nil {
		slog.Error("Failed to Create OPTIONS MOCK", "err: ", err)
		return nil, nil, err
	}

	methodResponse, err := apigateway.NewMethodResponse(mod.Ctx, fmt.Sprintf("%s-optionsMethodResponse", input.Endpoint.Name), &apigateway.MethodResponseArgs{
		RestApi:    input.RestApi.ID(),
		ResourceId: input.BaseResource.ID(),
		HttpMethod: pulumi.String("OPTIONS"),
		StatusCode: pulumi.String("200"),
		ResponseModels: pulumi.StringMap{
			"application/json": pulumi.String("Empty"),
		},
		ResponseParameters: pulumi.BoolMap{
			"method.response.header.Access-Control-Allow-Headers": pulumi.Bool(true),
			"method.response.header.Access-Control-Allow-Methods": pulumi.Bool(true),
			"method.response.header.Access-Control-Allow-Origin":  pulumi.Bool(true),
		},
	}, pulumi.DependsOn([]pulumi.Resource{optionsMethod, integration}))
	if err != nil {
		slog.Error("Failed to Create MethodResponse OPTIONS", "err: ", err)
		return nil, nil, err
	}

	integrationResponse, err := apigateway.NewIntegrationResponse(mod.Ctx, fmt.Sprintf("%s-optionsIntegrationResponse", input.Endpoint.Name), &apigateway.IntegrationResponseArgs{
		RestApi:    input.RestApi.ID(),
		ResourceId: input.BaseResource.ID(),
		HttpMethod: pulumi.String("OPTIONS"),
		StatusCode: pulumi.String("200"),
		ResponseParameters: pulumi.StringMap{
			"method.response.header.Access-Control-Allow-Headers": pulumi.String("'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'"),
			"method.response.header.Access-Control-Allow-Methods": pulumi.String("'POST,OPTIONS,GET,PUT,PATCH,DELETE'"),
			"method.response.header.Access-Control-Allow-Origin":  pulumi.String("'*'"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{optionsMethod, integration, methodResponse}))
	if err != nil {
		slog.Error("Failed to Create IntegrationResponse OPTIONS", "err: ", err)
		return nil, nil, err
	}
	return integrationResponse, optionsMethod, nil
}

func (mod RESTModule) CreateMethodIntegration(input dto.CreateMethodIntegrationInput) error {
	var methodArgs = &apigateway.MethodArgs{
		RestApi:       input.ApiID,
		ResourceId:    input.RootResourceID,
		HttpMethod:    pulumi.String(input.HttpMethod),
		Authorization: pulumi.String("NONE"),
	}
	if input.AuthorizerId != nil {
		methodArgs = &apigateway.MethodArgs{
			RestApi:       input.ApiID,
			ResourceId:    input.RootResourceID,
			HttpMethod:    pulumi.String(input.HttpMethod),
			Authorization: pulumi.String("CUSTOM"),
			AuthorizerId:  *input.AuthorizerId,
		}
	}
	createdMethod, err := apigateway.NewMethod(mod.Ctx, fmt.Sprintf("%s-method", input.Name), methodArgs)
	if err != nil {
		return err
	}

	_, err = apigateway.NewIntegration(mod.Ctx, fmt.Sprintf("%s-integration", input.Name), &apigateway.IntegrationArgs{
		RestApi:               input.ApiID,
		ResourceId:            input.RootResourceID,
		HttpMethod:            createdMethod.HttpMethod,
		IntegrationHttpMethod: pulumi.String("POST"),
		Type:                  pulumi.String("AWS_PROXY"),
		Uri:                   input.TargetLambdaArn,
	})
	if err != nil {
		return err
	}
	return nil
}

func (mod RESTModule) CreateStage(input dto.CreateStageInput) (*apigateway.Stage, error) {
	stage, err := apigateway.NewStage(mod.Ctx, fmt.Sprintf("%s-stage", input.BaseName), &apigateway.StageArgs{
		RestApi:     input.RestApi.ID(),
		Deployment:  input.Deploy.ID(),
		StageName:   pulumi.String(input.Name),
		Description: pulumi.String(input.BaseName),
	}, pulumi.DependsOn(input.AllOptions))
	if err != nil {
		slog.Error("Failed to Create API Stage", "err", err)
		return nil, err
	}

	_, err = apigateway.NewMethodSettings(mod.Ctx, "methodSettings", &apigateway.MethodSettingsArgs{
		RestApi:   input.RestApi.ID(),
		StageName: stage.StageName,
		Settings: apigateway.MethodSettingsSettingsArgs{
			LoggingLevel:     pulumi.String("INFO"),
			MetricsEnabled:   pulumi.Bool(true),
			DataTraceEnabled: pulumi.Bool(true),
		},
		MethodPath: pulumi.String("*/*"),
	})
	if err != nil {
		slog.Error("Failed to Create API Stage Method settings", "err", err)
		return nil, err
	}

	return stage, nil
}

func (mod RESTModule) CreateAuthorizer(input dto.CreateAuthorizerInput) (*apigateway.Authorizer, error) {
	invokeRole, err := iam.NewRole(mod.Ctx, fmt.Sprintf("%s-invocation-role", input.BaseName), &iam.RoleArgs{
		Name:             pulumi.String(fmt.Sprintf("%s-invocation-role", input.BaseName)),
		Path:             pulumi.String("/"),
		Tags:             input.Tags,
		AssumeRolePolicy: pulumi.String(policy.IAM_APIGW_ASSUME_ROLE),
	})
	if err != nil {
		slog.Error("error creating IAM role", "err:", err)
		return nil, err
	}

	_, err = iam.NewRolePolicy(mod.Ctx, fmt.Sprintf("%s-invoke-policy", input.BaseName), &iam.RolePolicyArgs{
		Role: invokeRole.ID(),
		Policy: pulumi.All(input.LambdaAuth.Arn).ApplyT(func(all []any) (string, error) {
			lambdaArn := all[0].(string)
			doc := fmt.Sprintf(string(policy.IAM_LAMBDA_INVOKE_ROLE), lambdaArn)
			return doc, nil
		}).(pulumi.StringOutput),
	})
	if err != nil {
		slog.Error("error creating IAM policy", "err:", err)
		return nil, err
	}

	authorizer, err := apigateway.NewAuthorizer(mod.Ctx, fmt.Sprintf("%s-authorizer", input.BaseName), &apigateway.AuthorizerArgs{
		Name:                         pulumi.StringPtr(fmt.Sprintf("%s-authorizer", input.BaseName)),
		RestApi:                      input.RestApi.ID(),
		AuthorizerUri:                input.LambdaAuth.InvokeArn,
		AuthorizerCredentials:        invokeRole.Arn,
		AuthorizerResultTtlInSeconds: pulumi.Int(0),
		IdentitySource:               pulumi.String(fmt.Sprintf("method.request.header.%s", input.IdentitySource)),
		Type:                         pulumi.String("TOKEN"),
	})
	if err != nil {
		slog.Error("Failed to Create API Authorizer", "err: ", err)
		return nil, err
	}

	_, err = lambda.NewPermission(mod.Ctx, fmt.Sprintf("%s-%s-lambda-permission", input.BaseName, "authorizer"), &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Function:  input.LambdaAuth.Name,
		Principal: pulumi.String("apigateway.amazonaws.com"),
		SourceArn: pulumi.Sprintf("arn:aws:execute-api:%s:%s:%s/authorizers/%s", input.Region, input.AccountId, input.RestApi.ID(), authorizer.ID().ToStringOutput()),
	})
	if err != nil {
		slog.Error("Failed to Create Authorizer apiGW permission", "err: ", err)
		return nil, err
	}
	return authorizer, nil
}

func validateUniqueEndpoints(endpoints []dto.Endpoints) error {
	seen := make(map[string]struct{})
	for _, ep := range endpoints {
		key := ep.Name + "::" + ep.Path
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate endpoint found: Name='%s', Path='%s'", ep.Name, ep.Path)
		}
		seen[key] = struct{}{}
	}
	return nil
}
