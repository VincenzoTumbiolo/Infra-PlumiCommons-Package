package dto

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type LambdaType string

const (
	LambdaTypeEFS LambdaType = "EFS"
	LambdaTypeS3  LambdaType = "S3"
)

type LambdaArgs struct {
	Name                         string
	Runtime                      string
	Handler                      string
	Architecture                 string
	MemorySize                   int
	Timeout                      int
	ReservedConcurrentExecutions pulumi.IntPtrInput
	Layers                       []string
	Description                  *string
	Publish                      pulumi.BoolPtrInput
	Tags                         pulumi.StringMap

	// Statements policy
	Statements []iam.GetPolicyDocumentStatementArgs

	// Source code
	BuildCommand   string
	WorkingDir     string
	OutputPath     string
	SourceCodeHash string
	ProjectPrefix  string

	// Environment variables
	Environments *lambda.FunctionEnvironmentArgs

	// VPC
	VpcConfig *lambda.FunctionVpcConfigArgs

	// Tracing
	TracingMode   *string
	TracingConfig *lambda.FunctionTracingConfigArgs

	// Dead Letter
	DeadLetterTargetArn *string
	DeadLetterConfig    *lambda.FunctionDeadLetterConfigArgs
}

type LambdaEFSArgs struct {
	SourceZipPath pulumi.ArchiveInput
}

type LambdaS3Args struct {
	S3Bucket string
	S3Key    string
}

type LambdaS3Input struct {
	LambdaArgs
	LambdaS3Args
}

type RESTLambdaArgs struct {
	LambdaArgs
	*LambdaEFSArgs
	*LambdaS3Args
	LambdaType LambdaType
}
