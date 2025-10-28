package lambda

import (
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	lambda_core "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/internal/aws/lambda/core"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateLambdaEFS(ctx *pulumi.Context, args dto.LambdaArgs, roleArn pulumi.StringInput) (*lambda.Function, error) {

	archive, err := lambda_core.BuildSourceZip(ctx, args.BuildCommand, args.WorkingDir, args.OutputPath)
	if err != nil {
		return nil, err
	}
	fnArgs := &lambda.FunctionArgs{
		Name:          pulumi.String(args.Name),
		Description:   pulumi.StringPtrFromPtr(args.Description),
		Role:          roleArn,
		Runtime:       pulumi.StringPtr(args.Runtime),
		Handler:       pulumi.StringPtr(args.Handler),
		MemorySize:    pulumi.IntPtr(args.MemorySize),
		Timeout:       pulumi.IntPtr(args.Timeout),
		Code:          archive.(pulumi.ArchiveInput),
		Architectures: pulumi.StringArray{pulumi.String(args.Architecture)},
		Layers:        pulumi.ToStringArray(args.Layers),
		VpcConfig:     args.VpcConfig,
		Tags:          args.Tags,
	}

	if args.Description != nil {
		fnArgs.Description = pulumi.StringPtrFromPtr(args.Description)
	}
	if args.Publish != nil {
		fnArgs.Publish = args.Publish
	}
	if args.ReservedConcurrentExecutions != nil {
		fnArgs.ReservedConcurrentExecutions = args.ReservedConcurrentExecutions
	}
	if args.Environments != nil {
		fnArgs.Environment = args.Environments
	}
	if args.VpcConfig != nil {
		fnArgs.VpcConfig = args.VpcConfig
	}
	if args.TracingMode != nil {
		fnArgs.TracingConfig = &lambda.FunctionTracingConfigArgs{
			Mode: pulumi.String(*args.TracingMode),
		}
	}
	if args.DeadLetterTargetArn != nil {
		fnArgs.DeadLetterConfig = &lambda.FunctionDeadLetterConfigArgs{
			TargetArn: pulumi.String(*args.DeadLetterTargetArn),
		}
	}

	resp, err := lambda.NewFunction(ctx, args.Name, fnArgs)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
