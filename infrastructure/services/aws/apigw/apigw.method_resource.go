package apigw_core

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func MethodResource(ctx *pulumi.Context, id pulumi.IDOutput, name string, rootResourceId string, pathPart string, httpMethod string) (*apigateway.Resource, *apigateway.Method, error) {
	resource, err := apigateway.NewResource(ctx, fmt.Sprintf("%s-resource", name), &apigateway.ResourceArgs{
		RestApi:  id,
		ParentId: pulumi.String(rootResourceId),
		PathPart: pulumi.String(pathPart),
	})
	if err != nil {
		return nil, nil, err
	}

	method, err := apigateway.NewMethod(ctx, fmt.Sprintf("%s-method", name), &apigateway.MethodArgs{
		RestApi:       id,
		ResourceId:    resource.ID(),
		HttpMethod:    pulumi.String(httpMethod),
		Authorization: pulumi.String("NONE"),
	})
	if err != nil {
		return resource, nil, err
	}

	return resource, method, nil
}
