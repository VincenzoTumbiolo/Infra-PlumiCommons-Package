package lambda

import (
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateLambdaS3(ctx *pulumi.Context, args dto.LambdaS3Input, roleArn pulumi.StringInput) (*lambda.Function, error) {
	fnArgs := &lambda.FunctionArgs{
		S3Bucket:                     pulumi.StringPtr(args.S3Bucket),
		S3Key:                        pulumi.StringPtr(args.S3Key),
		SourceCodeHash:               pulumi.StringPtr(args.SourceCodeHash),
		Name:                         pulumi.String(args.Name),
		Description:                  pulumi.StringPtrFromPtr(args.Description),
		Role:                         roleArn,
		Runtime:                      pulumi.StringPtr(args.Runtime),
		Handler:                      pulumi.StringPtr(args.Handler),
		MemorySize:                   pulumi.IntPtr(args.MemorySize),
		Timeout:                      pulumi.IntPtr(args.Timeout),
		ReservedConcurrentExecutions: args.ReservedConcurrentExecutions,
		Architectures:                pulumi.StringArray{pulumi.String(args.Architecture)},
		Layers:                       pulumi.ToStringArray(args.Layers),
		Environment:                  args.Environments,
		TracingConfig:                args.TracingConfig,
		VpcConfig:                    args.VpcConfig,
		DeadLetterConfig:             args.DeadLetterConfig,
		Tags:                         args.Tags,
	}

	if args.Description != nil {
		fnArgs.Description = pulumi.StringPtrFromPtr(args.Description)
	}
	if args.ReservedConcurrentExecutions != nil {
		fnArgs.ReservedConcurrentExecutions = args.ReservedConcurrentExecutions
	}
	if args.Environments != nil {
		fnArgs.Environment = args.Environments
	}
	if args.TracingConfig != nil {
		fnArgs.TracingConfig = args.TracingConfig
	}
	if args.VpcConfig != nil {
		fnArgs.VpcConfig = args.VpcConfig
	}
	if args.DeadLetterConfig != nil {
		fnArgs.DeadLetterConfig = args.DeadLetterConfig
	}

	return lambda.NewFunction(ctx, args.Name, fnArgs)
}
