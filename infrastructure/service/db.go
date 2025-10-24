package service

import (
	"fmt"

	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (mod ServiceModule) CreatePostgresCluster(name string, args *dto.PostgresClusterArgs) (*rds.Cluster, error) {

	// Subnet Group
	subnetGroup, err := rds.NewSubnetGroup(mod.Ctx, fmt.Sprintf("%s-db-subnet-group", name), &rds.SubnetGroupArgs{
		Name:      pulumi.String(fmt.Sprintf("%s-db-subnet-group", name)),
		SubnetIds: pulumi.ToStringArray(args.SubnetIds),
		Tags:      mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}

	// Cluster
	cluster, err := rds.NewCluster(mod.Ctx, fmt.Sprintf("%s-db", name), &rds.ClusterArgs{
		ClusterIdentifier:       pulumi.StringPtr(fmt.Sprintf("%s-db", name)),
		Engine:                  pulumi.String(args.Engine),
		Port:                    pulumi.IntPtrFromPtr(args.Port),
		EngineMode:              pulumi.StringPtrFromPtr(args.EngineMode),
		EngineVersion:           pulumi.StringPtrFromPtr(args.EngineVersion),
		DatabaseName:            pulumi.StringPtrFromPtr(args.DbName),
		MasterUsername:          pulumi.StringPtrFromPtr(args.MasterUsername),
		MasterPassword:          pulumi.StringPtrFromPtr(args.MasterPassword),
		DbSubnetGroupName:       subnetGroup.Name,
		VpcSecurityGroupIds:     args.SecurityGroupIds,
		BackupRetentionPeriod:   pulumi.IntPtrFromPtr(args.BackupRetentionPeriod),
		SkipFinalSnapshot:       pulumi.BoolPtr(args.SkipFinalSnapshot),
		FinalSnapshotIdentifier: pulumi.StringPtr(fmt.Sprintf("%s-db-final-snapshot", name)),
		DeleteAutomatedBackups:  pulumi.BoolPtr(args.DeleteAutomatedBackups),
		DeletionProtection:      pulumi.BoolPtr(args.DeletionProtection),
		Tags:                    mod.DefaultTags,
	}, pulumi.Protect(true))
	if err != nil {
		return nil, err
	}

	// Instances, based on ClusterSize
	for i := 0; i < args.ClusterSize; i++ {
		_, err := rds.NewClusterInstance(mod.Ctx, fmt.Sprintf("%s-instance-%d", name, i), &rds.ClusterInstanceArgs{
			Identifier:         pulumi.String(fmt.Sprintf("%s-instance-%d", name, i)),
			ClusterIdentifier:  cluster.ID(),
			InstanceClass:      pulumi.String(args.ClusterInstanceClass),
			Engine:             rds.EngineType(args.Engine),
			PubliclyAccessible: pulumi.BoolPtrFromPtr(args.PubliclyAccessible),
			Tags:               mod.DefaultTags,
		}, pulumi.DependsOn([]pulumi.Resource{cluster}))
		if err != nil {
			return nil, err
		}
	}

	return cluster, nil
}
