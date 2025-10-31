package load_balancer

import (
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateTargetGroup(ctx *pulumi.Context, in dto.TargetGroupInput) (*lb.TargetGroup, error) {
	// tags fallback
	tags := in.Tags
	if tags == nil {
		tags = pulumi.StringMap{}
	}

	tg, err := lb.NewTargetGroup(ctx, in.Name, &lb.TargetGroupArgs{
		Name:       pulumi.String(in.Name),
		Port:       pulumi.Int(in.Port),
		TargetType: pulumi.String(in.TargetType),
		Protocol:   pulumi.String(in.Protocol),
		VpcId:      pulumi.String(in.VpcId),
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Path:               pulumi.String(in.HealthCheckPath),
			Port:               pulumi.String(in.HealthCheckPort),
			Protocol:           pulumi.String(in.HealthCheckProtocol),
			HealthyThreshold:   pulumi.Int(in.HealthCheckHealthyThreshold),
			UnhealthyThreshold: pulumi.Int(in.HealthCheckUnhealthyThreshold),
			Matcher:            pulumi.String(in.HealthCheckMatcher),
		},
		Tags: tags,
	})
	if err != nil {
		return nil, err
	}

	return tg, nil
}
