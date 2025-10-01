package network

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	dto "github.com/vtechstudio/infra-pulumicommons-package/infrastructure/dto/aws"
	"github.com/vtechstudio/infra-pulumicommons-package/infrastructure/mappers"
)

func CreateSecurityGroup(ctx *pulumi.Context, name string, args *dto.SecurityGroupArgs) (*ec2.SecurityGroup, error) {
	sg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId:       pulumi.StringPtrFromPtr(args.VpcID),
		Description: pulumi.StringPtrFromPtr(args.Description),
		Tags:        args.Tags,
		Ingress:     mappers.MapRulesToIngress(args.Ingress),
		Egress:      mappers.MapRulesToEgress(args.Egress),
	})
	if err != nil {
		return nil, fmt.Errorf("creating security group: %w", err)
	}

	return sg, nil
}
