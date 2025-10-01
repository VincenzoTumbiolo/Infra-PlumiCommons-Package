package dto

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PostgresClusterArgs struct {
	Tags                   pulumi.StringMap
	SubnetIds              []string
	SecurityGroupIds       pulumi.StringArray
	MasterUsername         *string
	MasterPassword         *string
	Port                   *int
	ClusterInstanceClass   string
	Engine                 string
	EngineMode             *string
	EngineVersion          *string
	DbName                 *string
	BackupRetentionPeriod  *int
	SkipFinalSnapshot      bool
	ClusterSize            int
	PubliclyAccessible     *bool
	DeleteAutomatedBackups bool
	DeletionProtection     bool

	Ingress ec2.SecurityGroupIngressArray
	Egress  ec2.SecurityGroupEgressArray
}
