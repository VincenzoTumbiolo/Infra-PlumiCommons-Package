package dto

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type SecurityGroupRule struct {
	Protocol    string
	FromPort    int
	ToPort      int
	CidrBlock   string
	Description string
}

type SecurityGroupArgs struct {
	Description *string
	VpcID       *string
	Ingress     []SecurityGroupRule
	Egress      []SecurityGroupRule
	Tags        pulumi.StringMapInput
}
