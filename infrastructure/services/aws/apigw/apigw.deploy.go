package apigw_core

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Deploy(ctx *pulumi.Context, id pulumi.IDOutput, name string) (*apigateway.Deployment, error) {
	deployment, err := apigateway.NewDeployment(ctx, fmt.Sprintf("%s-deployment", name), &apigateway.DeploymentArgs{
		RestApi: id,
	})
	if err != nil {
		return nil, err
	}

	return deployment, nil
}
