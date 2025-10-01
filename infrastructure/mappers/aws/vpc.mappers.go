package mappers

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func GetVpcConfigArgs(
	subnetIds []string,
	securityGroupIds pulumi.StringArray,
) *lambda.FunctionVpcConfigArgs {
	return &lambda.FunctionVpcConfigArgs{
		SubnetIds:        pulumi.ToStringArray(subnetIds),
		SecurityGroupIds: securityGroupIds,
	}
}
