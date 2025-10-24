package service

import (
	"errors"
	"fmt"

	policy "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/aws"
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	lambda_services "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/internal/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func (mod ServiceModule) CreateLambda(args *dto.ServiceLambdaArgs) (*lambda.Function, error) {
	args.LambdaArgs.Tags = mod.DefaultTags

	var statements = mod.Policies.Build(
		policy.StatementSpec{
			Groups:    []policy.PolicyGroup{policy.CLOUDWATCH_LOGS},
			Resources: policy.AllResources(),
		},
		policy.StatementSpec{
			Groups:    []policy.PolicyGroup{policy.EC2_ENI_LIFECYCLE},
			Resources: policy.AllResources(),
		},
	)

	for _, v := range args.LambdaArgs.Statements {
		statements = append(statements, v)
	}
	// Policy document
	getDumperPolicyDoc := iam.GetPolicyDocumentOutput(mod.Ctx, iam.GetPolicyDocumentOutputArgs{
		Statements: statements,
	})

	// IAM Role
	getDumperRole, err := iam.NewRole(mod.Ctx, fmt.Sprintf("%s-role", args.LambdaArgs.Name), &iam.RoleArgs{
		Name:             pulumi.String(fmt.Sprintf("%s-role", args.LambdaArgs.Name)),
		AssumeRolePolicy: pulumi.String(policy.IAM_LAMBDA_ASSUME_ROLE),
	})
	if err != nil {
		return nil, err
	}

	// IAM Role Policy
	_, err = iam.NewRolePolicy(mod.Ctx, fmt.Sprintf("%s-role", args.LambdaArgs.Name), &iam.RolePolicyArgs{
		Name:   pulumi.String(fmt.Sprintf("%s-policy", args.LambdaArgs.Name)),
		Role:   getDumperRole.ID(),
		Policy: getDumperPolicyDoc.Json(),
	})
	if err != nil {
		return nil, err
	}

	var fn *lambda.Function
	switch args.LambdaType {
	case dto.LambdaTypeEFS:
		fn, err = lambda_services.CreateLambdaEFS(mod.Ctx, args.LambdaArgs, getDumperRole.Arn)
	case dto.LambdaTypeS3:
		fn, err = lambda_services.CreateLambdaS3(mod.Ctx, dto.LambdaS3Input{
			LambdaArgs:   args.LambdaArgs,
			LambdaS3Args: *args.LambdaS3Args,
		}, getDumperRole.Arn)
	default:
		fn, err = nil, errors.New("unsupported lambda type")
	}
	if err != nil {
		return nil, err
	}

	// Log group associato alla Lambda
	_, err = cloudwatch.NewLogGroup(mod.Ctx, fmt.Sprintf("%s-log-group", args.LambdaArgs.Name), &cloudwatch.LogGroupArgs{
		Name:            pulumi.String(fmt.Sprintf("/aws/lambda/%s", args.LambdaArgs.Name)),
		RetentionInDays: pulumi.Int(30),
		Tags:            mod.DefaultTags,
	})
	if err != nil {
		return nil, err
	}

	return fn, nil
}
