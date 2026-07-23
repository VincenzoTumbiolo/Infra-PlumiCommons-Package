package load_balancer

import (
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// CreateService crea l'ALB/NLB con gli stessi valori del Terraform dato.
// - internal = true
// - enable_deletion_protection = false
// - access_logs abilitati su LogBucket con prefix = LbName
// - security_groups solo se LbType == "application"
func CreateService(ctx *pulumi.Context, in dto.LoadBalancerInput) (*lb.LoadBalancer, error) {
	// security_groups: solo per "application"
	var sgs pulumi.StringArray
	if in.LbType == "application" && in.LbSecurityGroupId != nil {
		sgs = pulumi.StringArray{*in.LbSecurityGroupId}
	}

	// access_logs
	accessLogs := &lb.LoadBalancerAccessLogsArgs{
		Bucket:  pulumi.String(in.LogBucket),
		Prefix:  pulumi.StringPtr(in.LbName),
		Enabled: pulumi.Bool(true),
	}

	// tags fallback
	tags := in.Tags
	if tags == nil {
		tags = pulumi.StringMap{}
	}

	lbRes, err := lb.NewLoadBalancer(ctx, in.LbName, &lb.LoadBalancerArgs{
		Name:                     pulumi.String(in.LbName),
		Internal:                 pulumi.Bool(true),
		LoadBalancerType:         pulumi.String(in.LbType),
		SecurityGroups:           sgs,                                  // nil se non applicabile
		Subnets:                  pulumi.ToStringArray(in.LbSubnetIds), // obbligatorio
		EnableDeletionProtection: pulumi.Bool(false),                   // come Terraform
		AccessLogs:               accessLogs,
		Tags:                     tags,
	})
	if err != nil {
		return nil, err
	}

	return lbRes, nil
}
