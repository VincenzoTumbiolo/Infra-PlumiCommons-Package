package dto

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PostgresClusterArgs struct {
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

	Serverlessv2ScalingConfiguration *rds.ClusterServerlessv2ScalingConfigurationArgs

	Tags pulumi.StringMap
}

type PostgresInstanceArgs struct {
	Engine                 string
	EngineVersion          *string
	InstanceClass          string
	AllocatedStorage       *int
	MaxAllocatedStorage    *int
	StorageType            *string
	DbName                 *string
	MasterUsername         *string
	MasterPassword         *string
	Port                   *int
	SubnetIds              []string
	SecurityGroupIds       pulumi.StringArrayInput
	PubliclyAccessible     *bool
	BackupRetentionPeriod  *int
	SkipFinalSnapshot      *bool
	DeleteAutomatedBackups *bool
	DeletionProtection     *bool
	Tags                   pulumi.StringMap
}
