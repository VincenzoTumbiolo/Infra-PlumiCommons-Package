package service

import (
	"encoding/json"
	"fmt"

	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/appautoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (mod ServiceModule) CreateServiceStandAlone(ctx *pulumi.Context, baseName string, in dto.ECSInput) (*ecs.Service, error) {
	// Cluster
	clusterName := pulumi.String(baseName + "-ecs-cluster")
	if _, err := ecs.NewCluster(ctx, baseName+"-cluster", &ecs.ClusterArgs{
		Name: clusterName,
		Tags: mod.DefaultTags,
	}); err != nil {
		return nil, err
	}

	// CloudWatch Log Group
	retention := 14
	if in.LogRetentionDays > 0 {
		retention = in.LogRetentionDays
	}
	lg, err := cloudwatch.NewLogGroup(ctx, baseName+"-lg", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(in.LogGroupName),
		RetentionInDays: pulumi.Int(retention),
		Tags:            mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}

	// IAM roles (execution + task)
	execRole, err := iam.NewRole(ctx, baseName+"-exec-role", &iam.RoleArgs{
		Name: pulumi.String(baseName + "-ecsTaskExecutionRole"),
		AssumeRolePolicy: pulumi.String(`{
		  "Version":"2012-10-17",
		  "Statement":[{"Effect":"Allow","Principal":{"Service":"ecs-tasks.amazonaws.com"},"Action":"sts:AssumeRole"}]
		}`),
		Tags: mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}
	if _, err = iam.NewRolePolicyAttachment(ctx, baseName+"-exec-pol", &iam.RolePolicyAttachmentArgs{
		Role:      execRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	}); err != nil {
		return nil, err
	}

	taskRole, err := iam.NewRole(ctx, baseName+"-task-role", &iam.RoleArgs{
		Name: pulumi.String(baseName + "-ecsTaskRole"),
		AssumeRolePolicy: pulumi.String(`{
		  "Version":"2012-10-17",
		  "Statement":[{"Effect":"Allow","Principal":{"Service":"ecs-tasks.amazonaws.com"},"Action":"sts:AssumeRole"}]
		}`),
		Tags: mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}
	if _, err = iam.NewRolePolicy(ctx, baseName+"-task-logs", &iam.RolePolicyArgs{
		Name: pulumi.String(baseName + "-log-policy"),
		Role: taskRole.Name,
		Policy: pulumi.String(fmt.Sprintf(`{
		  "Version":"2012-10-17",
		  "Statement":[
		    {"Sid":"Logs","Effect":"Allow",
		      "Action":["logs:CreateLogGroup","logs:CreateLogStream","logs:PutLogEvents","logs:DescribeLogStreams"],
		      "Resource":"arn:aws:logs:%s:%s:*"}
		  ]}`, in.Region, in.AccountID)),
	}); err != nil {
		return nil, err
	}
	if in.AttachS3FullAccess {
		_, _ = iam.NewRolePolicyAttachment(ctx, baseName+"-task-s3", &iam.RolePolicyAttachmentArgs{
			Role:      taskRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonS3FullAccess"),
		})
	}
	if in.CustomInlinePolicyJSON != nil && *in.CustomInlinePolicyJSON != "" {
		_, _ = iam.NewRolePolicy(ctx, baseName+"-task-custom", &iam.RolePolicyArgs{
			Name:   pulumi.String(baseName + "-custom-policy"),
			Role:   taskRole.Name,
			Policy: pulumi.String(*in.CustomInlinePolicyJSON),
		})
	}

	// Container defs
	appContainer := map[string]any{
		"name":        baseName + "-container",
		"image":       in.EcrImageUrl,
		"essential":   true,
		"environment": in.EnvVars, // []EnvVar
		"portMappings": []dto.PortMapping{{
			Protocol: "tcp", ContainerPort: in.ContainerPort, HostPort: in.HostPort,
		}},
		"logConfiguration": map[string]any{
			"logDriver": "awslogs",
			"options": map[string]string{
				"awslogs-group":         in.LogGroupName,
				"awslogs-region":        in.Region,
				"awslogs-stream-prefix": "ecs",
			},
		},
	}
	containers := []any{appContainer}
	if in.DDog.Enable {
		containers = append(containers, map[string]any{
			"name":         baseName + "-ddog",
			"image":        in.DDog.Image,
			"environment":  in.DDog.Env,
			"portMappings": in.DDog.PortMappings,
			"essential":    false,
			"logConfiguration": map[string]any{
				"logDriver": "awslogs",
				"options": map[string]string{
					"awslogs-group":         in.LogGroupName,
					"awslogs-region":        in.Region,
					"awslogs-stream-prefix": "ddog",
				},
			},
		})
	}
	cdef, _ := json.Marshal(containers)

	// Task Definition (Fargate)
	td, err := ecs.NewTaskDefinition(ctx, baseName+"-taskdef", &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(baseName + "-task"),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.ToStringArray([]string{"FARGATE"}),
		Cpu:                     pulumi.String(in.TaskCPU),
		Memory:                  pulumi.String(in.TaskMemory),
		ExecutionRoleArn:        execRole.Arn,
		TaskRoleArn:             taskRole.Arn,
		ContainerDefinitions:    pulumi.String(cdef),
		Tags:                    mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}

	// Service (pubblico, NO LB)
	var subnets pulumi.StringArray
	for _, s := range in.SubnetIds {
		subnets = append(subnets, s)
	}
	var sgs pulumi.StringArray
	for _, sg := range in.SecurityGroupIds {
		sgs = append(sgs, sg)
	}

	svc, err := ecs.NewService(ctx, baseName+"-svc", &ecs.ServiceArgs{
		Name:           pulumi.String(fmt.Sprintf("%s-service", baseName)),
		Cluster:        clusterName,
		TaskDefinition: td.Arn,
		LaunchType:     pulumi.String("FARGATE"),
		DesiredCount:   pulumi.Int(in.DesiredCount),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(in.AssignPublicIp),
			Subnets:        subnets,
			SecurityGroups: sgs,
		},
		Tags: mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}

	// Autoscaling DesiredCount
	resID := pulumi.Sprintf("service/%s/%s", clusterName, svc.Name)
	asgTarget, err := appautoscaling.NewTarget(ctx, baseName+"-as-target", &appautoscaling.TargetArgs{
		MaxCapacity:       pulumi.Int(in.MaxCapacity),
		MinCapacity:       pulumi.Int(in.MinCapacity),
		ResourceId:        resID,
		ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
		ServiceNamespace:  pulumi.String("ecs"),
	})
	if err != nil {
		return nil, err
	}
	_, _ = appautoscaling.NewPolicy(ctx, baseName+"-as-mem", &appautoscaling.PolicyArgs{
		PolicyType:        pulumi.String("TargetTrackingScaling"),
		ResourceId:        asgTarget.ResourceId,
		ScalableDimension: asgTarget.ScalableDimension,
		ServiceNamespace:  asgTarget.ServiceNamespace,
		TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
			PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
				PredefinedMetricType: pulumi.String("ECSServiceAverageMemoryUtilization"),
			},
			TargetValue: pulumi.Float64(in.TargetMemory),
		},
	})
	_, _ = appautoscaling.NewPolicy(ctx, baseName+"-as-cpu", &appautoscaling.PolicyArgs{
		PolicyType:        pulumi.String("TargetTrackingScaling"),
		ResourceId:        asgTarget.ResourceId,
		ScalableDimension: asgTarget.ScalableDimension,
		ServiceNamespace:  asgTarget.ServiceNamespace,
		TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
			PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
				PredefinedMetricType: pulumi.String("ECSServiceAverageCPUUtilization"),
			},
			TargetValue: pulumi.Float64(in.TargetCPU),
		},
	})

	// Log subscription → Lambda (opzionale)
	if in.LogSubscriptionLambdaName != nil && in.LogSubscriptionLambdaArn != nil {
		perm, _ := lambda.NewPermission(ctx, baseName+"-lg-lambda-perm", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  pulumi.String(*in.LogSubscriptionLambdaName),
			Principal: pulumi.String(fmt.Sprintf("logs.%s.amazonaws.com", in.Region)),
			SourceArn: pulumi.Sprintf("%s:*", lg.Arn),
		})
		_, _ = cloudwatch.NewLogSubscriptionFilter(ctx, baseName+"-lg-sub", &cloudwatch.LogSubscriptionFilterArgs{
			Name:           pulumi.String(baseName + "-lambda-subscription"),
			DestinationArn: pulumi.String(*in.LogSubscriptionLambdaArn),
			FilterPattern:  pulumi.String(""),
			LogGroup:       lg.Name,
		}, pulumi.DependsOn([]pulumi.Resource{perm}))
	}

	// Route53 record (OPZIONALE)
	if in.HostedZoneId != nil && *in.HostedZoneId != "" && in.RecordName != nil && *in.RecordName != "" && in.BackendPublicIp != nil && *in.BackendPublicIp != "" {
		ttl := 60
		if in.RecordTTL != nil && *in.RecordTTL > 0 {
			ttl = *in.RecordTTL
		}
		// Se RecordName è relativo (es. "api"), Route53 lo normalizza con il nome della zona.
		_, err := route53.NewRecord(ctx, baseName+"-dns", &route53.RecordArgs{
			ZoneId: pulumi.String(*in.HostedZoneId),
			Name:   pulumi.String(*in.RecordName), // "api" o "api.miodominio.com"
			Type:   pulumi.String("A"),
			Ttl:    pulumi.Int(ttl),
			Records: pulumi.ToStringArray([]string{
				*in.BackendPublicIp,
			}),
		})
		if err != nil {
			return nil, err
		}
	} else {
		ctx.Log.Info("[DNS] Skipping Route53 A-record: missing HostedZoneId/RecordName/BackendPublicIp", nil)
	}

	return svc, nil
}

func (mod ServiceModule) CreateService(ctx *pulumi.Context, baseName string, in dto.ECSInput) error {
	// Cluster
	clusterName := pulumi.String(baseName + "-ecs-cluster")
	if _, err := ecs.NewCluster(ctx, baseName+"-cluster", &ecs.ClusterArgs{
		Name: clusterName,
		Tags: mod.DefaultTags,
	}); err != nil {
		return err
	}

	// CloudWatch Log Group (Logs)
	retention := 14
	if in.LogRetentionDays > 0 {
		retention = in.LogRetentionDays
	}
	lg, err := cloudwatch.NewLogGroup(ctx, baseName+"-lg", &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(in.LogGroupName),
		RetentionInDays: pulumi.Int(retention),
		Tags:            mod.DefaultTags,
	})
	if err != nil {
		return err
	}

	// IAM: Task Execution Role
	execRole, err := iam.NewRole(ctx, baseName+"-exec-role", &iam.RoleArgs{
		Name: pulumi.String(baseName + "-ecsTaskExecutionRole"),
		AssumeRolePolicy: pulumi.String(`{
		  "Version":"2012-10-17",
		  "Statement":[{"Effect":"Allow","Principal":{"Service":"ecs-tasks.amazonaws.com"},"Action":"sts:AssumeRole"}]
		}`),
		Tags: mod.DefaultTags,
	})
	if err != nil {
		return err
	}
	if _, err = iam.NewRolePolicyAttachment(ctx, baseName+"-exec-pol", &iam.RolePolicyAttachmentArgs{
		Role:      execRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	}); err != nil {
		return err
	}

	// IAM: Task Role
	taskRole, err := iam.NewRole(ctx, baseName+"-task-role", &iam.RoleArgs{
		Name: pulumi.String(baseName + "-ecsTaskRole"),
		AssumeRolePolicy: pulumi.String(`{
		  "Version":"2012-10-17",
		  "Statement":[{"Effect":"Allow","Principal":{"Service":"ecs-tasks.amazonaws.com"},"Action":"sts:AssumeRole"}]
		}`),
		Tags: mod.DefaultTags,
	})
	if err != nil {
		return err
	}
	// Policy minima per logs
	if _, err = iam.NewRolePolicy(ctx, baseName+"-task-logs", &iam.RolePolicyArgs{
		Name: pulumi.String(baseName + "-log-policy"),
		Role: taskRole.Name,
		Policy: pulumi.String(fmt.Sprintf(`{
		  "Version":"2012-10-17",
		  "Statement":[
		    {"Sid":"Logs","Effect":"Allow",
		      "Action":["logs:CreateLogGroup","logs:CreateLogStream","logs:PutLogEvents","logs:DescribeLogStreams"],
		      "Resource":"arn:aws:logs:%s:%s:*"}
		  ]}`, in.Region, in.AccountID)),
	}); err != nil {
		return err
	}
	// Extra S3 (opzionale)
	if in.AttachS3FullAccess {
		_, _ = iam.NewRolePolicyAttachment(ctx, baseName+"-task-s3", &iam.RolePolicyAttachmentArgs{
			Role:      taskRole.Name,
			PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonS3FullAccess"),
		})
	}
	// Inline custom (opzionale)
	if in.CustomInlinePolicyJSON != nil && *in.CustomInlinePolicyJSON != "" {
		_, _ = iam.NewRolePolicy(ctx, baseName+"-task-custom", &iam.RolePolicyArgs{
			Name:   pulumi.String(baseName + "-custom-policy"),
			Role:   taskRole.Name,
			Policy: pulumi.String(*in.CustomInlinePolicyJSON),
		})
	}

	// Container definitions
	appContainer := map[string]interface{}{
		"name":        baseName + "-container",
		"image":       in.EcrImageUrl,
		"essential":   true,
		"environment": in.EnvVars, // []EnvVar
		"portMappings": []dto.PortMapping{{
			Protocol: "tcp", ContainerPort: in.ContainerPort, HostPort: in.HostPort,
		}},
		"logConfiguration": map[string]interface{}{
			"logDriver": "awslogs",
			"options": map[string]string{
				"awslogs-group":         in.LogGroupName,
				"awslogs-region":        in.Region,
				"awslogs-stream-prefix": "ecs",
			},
		},
	}

	containers := []interface{}{appContainer}
	if in.DDog.Enable {
		dd := map[string]interface{}{
			"name":         baseName + "-ddog",
			"image":        in.DDog.Image,
			"environment":  in.DDog.Env,
			"portMappings": in.DDog.PortMappings,
			"essential":    false,
			"logConfiguration": map[string]interface{}{
				"logDriver": "awslogs",
				"options": map[string]string{
					"awslogs-group":         in.LogGroupName,
					"awslogs-region":        in.Region,
					"awslogs-stream-prefix": "ddog",
				},
			},
		}
		containers = append(containers, dd)
	}
	cdef, _ := json.Marshal(containers)

	// Task Definition (Fargate)
	td, err := ecs.NewTaskDefinition(ctx, baseName+"-taskdef", &ecs.TaskDefinitionArgs{
		Family:                  pulumi.String(baseName + "-task"),
		NetworkMode:             pulumi.String("awsvpc"),
		RequiresCompatibilities: pulumi.ToStringArray([]string{"FARGATE"}),
		Cpu:                     pulumi.String(in.TaskCPU),
		Memory:                  pulumi.String(in.TaskMemory),
		ExecutionRoleArn:        execRole.Arn,
		TaskRoleArn:             taskRole.Arn,
		ContainerDefinitions:    pulumi.String(cdef),
		Tags:                    mod.DefaultTags,
	})
	if err != nil {
		return err
	}

	// ECS Service (senza LB) — NB: IP pubblico del task può cambiare ai redeploy
	var subnets pulumi.StringArray
	for _, s := range in.SubnetIds {
		subnets = append(subnets, s)
	}
	var sgs pulumi.StringArray
	for _, sg := range in.SecurityGroupIds {
		sgs = append(sgs, sg)
	}

	svc, err := ecs.NewService(ctx, baseName+"-svc", &ecs.ServiceArgs{
		Name:           pulumi.String(fmt.Sprintf("%s-service", baseName)),
		Cluster:        clusterName,
		TaskDefinition: td.Arn,
		LaunchType:     pulumi.String("FARGATE"),
		DesiredCount:   pulumi.Int(in.DesiredCount),
		NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
			AssignPublicIp: pulumi.Bool(in.AssignPublicIp),
			Subnets:        subnets,
			SecurityGroups: sgs,
		},
		Tags: mod.DefaultTags,
	})
	if err != nil {
		return err
	}

	// Autoscaling: DesiredCount (CPU / Memory target tracking)
	resID := pulumi.Sprintf("service/%s/%s", clusterName, svc.Name)

	asgTarget, err := appautoscaling.NewTarget(ctx, baseName+"-as-target", &appautoscaling.TargetArgs{
		MaxCapacity:       pulumi.Int(in.MaxCapacity),
		MinCapacity:       pulumi.Int(in.MinCapacity),
		ResourceId:        resID,
		ScalableDimension: pulumi.String("ecs:service:DesiredCount"),
		ServiceNamespace:  pulumi.String("ecs"),
	})
	if err != nil {
		return err
	}

	_, _ = appautoscaling.NewPolicy(ctx, baseName+"-as-mem", &appautoscaling.PolicyArgs{
		PolicyType:        pulumi.String("TargetTrackingScaling"),
		ResourceId:        asgTarget.ResourceId,
		ScalableDimension: asgTarget.ScalableDimension,
		ServiceNamespace:  asgTarget.ServiceNamespace,
		TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
			PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
				PredefinedMetricType: pulumi.String("ECSServiceAverageMemoryUtilization"),
			},
			TargetValue: pulumi.Float64(in.TargetMemory),
		},
	})
	_, _ = appautoscaling.NewPolicy(ctx, baseName+"-as-cpu", &appautoscaling.PolicyArgs{
		PolicyType:        pulumi.String("TargetTrackingScaling"),
		ResourceId:        asgTarget.ResourceId,
		ScalableDimension: asgTarget.ScalableDimension,
		ServiceNamespace:  asgTarget.ServiceNamespace,
		TargetTrackingScalingPolicyConfiguration: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationArgs{
			PredefinedMetricSpecification: &appautoscaling.PolicyTargetTrackingScalingPolicyConfigurationPredefinedMetricSpecificationArgs{
				PredefinedMetricType: pulumi.String("ECSServiceAverageCPUUtilization"),
			},
			TargetValue: pulumi.Float64(in.TargetCPU),
		},
	})

	// Log subscription verso Lambda
	if in.LogSubscriptionLambdaName != nil && in.LogSubscriptionLambdaArn != nil {
		perm, _ := lambda.NewPermission(ctx, baseName+"-lg-lambda-perm", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  pulumi.String(*in.LogSubscriptionLambdaName),
			Principal: pulumi.String(fmt.Sprintf("logs.%s.amazonaws.com", in.Region)),
			SourceArn: pulumi.Sprintf("%s:*", lg.Arn),
		})
		_, _ = cloudwatch.NewLogSubscriptionFilter(ctx, baseName+"-lg-sub", &cloudwatch.LogSubscriptionFilterArgs{
			Name:           pulumi.String(baseName + "-lambda-subscription"),
			DestinationArn: pulumi.String(*in.LogSubscriptionLambdaArn),
			FilterPattern:  pulumi.String(""),
			LogGroup:       lg.Name,
		}, pulumi.DependsOn([]pulumi.Resource{perm}))
	}

	// return &dto.ECSOutput{
	// 	ClusterName:          clusterName.ToStringOutput(),
	// 	ServiceName:          svc.Name,
	// 	ServiceArn:           svc.Arn,
	// 	TaskDefinitionArn:    td.Arn,
	// 	TaskExecutionRoleArn: execRole.Arn,
	// 	TaskRoleArn:          taskRole.Arn,
	// 	LogGroupName:         lg.Name,
	// },
	return nil
}
