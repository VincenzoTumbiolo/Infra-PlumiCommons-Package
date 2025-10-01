package dto

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type RestAPIArgs struct {
	Name        string
	Description *string
	Tags        pulumi.StringMap
}

type MethodArgs struct {
	Name            string
	Region          string
	Method          string
	Path            string
	TargetLambdaArn pulumi.StringInput
}

type Methods struct {
	Name                     string
	HttpMethod               string
	TargetLambdaInvokeArn    pulumi.StringInput
	TargetLambdaFunctionName pulumi.Input
}

type Endpoints struct {
	Name    string
	Path    string
	Methods []Methods
}

type CreateRestAPI struct {
	BaseName  string
	Region    string
	AccountId string
}

type CreateRestAPIInput struct {
	CreateRestAPI

	Endpoints          []Endpoints
	LambdaAuth         *lambda.Function
	DomainName         string
	HttpsCertificateId string
	IdentitySource     string
	StageName          string
	Tags               pulumi.StringMapInput
}

type CreateEndpointsInput struct {
	CreateRestAPI

	Endpoints  []Endpoints
	RestApi    *apigateway.RestApi
	AllOptions []pulumi.Resource
	AuthID     *pulumi.StringPtrInput
}

type CreateMethodIntegrationInput struct {
	ApiID           pulumi.IDOutput
	Name            string
	RootResourceID  pulumi.IDOutput
	HttpMethod      string
	TargetLambdaArn pulumi.StringInput
	AuthorizerId    *pulumi.StringPtrInput
}

type CreateStageInput struct {
	CreateRestAPI
	RestApi    *apigateway.RestApi
	Deploy     *apigateway.Deployment
	Name       string
	Tags       pulumi.StringMapInput
	AllOptions []pulumi.Resource
}

type CreateAuthorizerInput struct {
	CreateRestAPI
	RestApi        *apigateway.RestApi
	Tags           pulumi.StringMapInput
	LambdaAuth     *lambda.Function
	IdentitySource string
}

type CreateOptionsInput struct {
	RestApi      *apigateway.RestApi
	BaseResource *apigateway.Resource
	Endpoint     Endpoints
	AllOptions   []pulumi.Resource
}
