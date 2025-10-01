package policy

import (
	"sort"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type IAMRoleArgs string

const (
	IAM_LAMBDA_ASSUME_ROLE IAMRoleArgs = `{
				"Version": "2012-10-17",
				"Statement": [{
					"Action": "sts:AssumeRole",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					},
					"Effect": "Allow"
				}]
			}`
	IAM_LAMBDA_INVOKE_ROLE IAMRoleArgs = `{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Action": "lambda:InvokeFunction",
						"Resource": "%s"
					}
				]
			}`
	IAM_APIGW_ASSUME_ROLE IAMRoleArgs = `{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Action": "sts:AssumeRole",
					"Principal": {
						"Service": "apigateway.amazonaws.com"
					},
					"Effect": "Allow",
				}
			]
		}`
)

type PolicyGroup string

const (
	// Logs & Monitoring
	CLOUDWATCH_LOGS PolicyGroup = "CLOUDWATCH_LOGS"

	// EC2 (Lambda in VPC: ENI lifecycle & IP assign)
	EC2_ENI_LIFECYCLE PolicyGroup = "EC2_ENI_LIFECYCLE"

	// Storage
	S3_MANAGE_FILE PolicyGroup = "S3_MANAGE_FILE"

	// Queues
	SQS_MANAGE PolicyGroup = "SQS_MANAGE"

	// DB
	DYNAMODB_RW PolicyGroup = "DYNAMODB_RW"

	// Compute
	LAMBDA_INVOKE PolicyGroup = "LAMBDA_INVOKE"

	// Notifications / Email
	SNS_SEND_EMAIL PolicyGroup = "SNS_SEND_EMAIL"
)

type PolicySet struct {
	groupActions map[PolicyGroup][]string
}

// Provide a default registry of common groups.
func DefaultPolicySet() *PolicySet {
	return &PolicySet{
		groupActions: map[PolicyGroup][]string{
			CLOUDWATCH_LOGS: {
				"logs:CreateLogGroup",
				"logs:CreateLogStream",
				"logs:PutLogEvents",
				"logs:DescribeLogStreams",
			},
			EC2_ENI_LIFECYCLE: {
				"ec2:CreateNetworkInterface",
				"ec2:DescribeNetworkInterfaces",
				"ec2:DeleteNetworkInterface",
				"ec2:AssignPrivateIpAddresses",
				"ec2:UnassignPrivateIpAddresses",
			},
			S3_MANAGE_FILE: {
				"s3:PutObject",
				"s3:GetObject",
				"s3:DeleteObject",
			},
			SQS_MANAGE: {
				"sqs:SendMessage",
				"sqs:ReceiveMessage",
				"sqs:DeleteMessage",
				"sqs:GetQueueAttributes",
				"sqs:GetQueueUrl",
				"sqs:ChangeMessageVisibility",
			},
			DYNAMODB_RW: {
				"dynamodb:PutItem",
				"dynamodb:GetItem",
				"dynamodb:UpdateItem",
				"dynamodb:DeleteItem",
				"dynamodb:Query",
				"dynamodb:Scan",
			},
			LAMBDA_INVOKE: {
				"lambda:InvokeFunction",
			},
			SNS_SEND_EMAIL: {
				"sns:Publish",
				"ses:SendEmail",
				"ses:SendRawEmail",
			},
		},
	}
}

// Allow projects to add/override groups if needed (optional).
func (ps *PolicySet) WithGroup(name PolicyGroup, actions ...string) *PolicySet {
	if ps.groupActions == nil {
		ps.groupActions = map[PolicyGroup][]string{}
	}
	cp := make([]string, len(actions))
	copy(cp, actions)
	ps.groupActions[name] = cp
	return ps
}

func (ps *PolicySet) actionsOf(g PolicyGroup) []string {
	src := ps.groupActions[g]
	out := make([]string, len(src))
	copy(out, src)
	return out
}

func (ps *PolicySet) merge(groups ...PolicyGroup) []string {
	set := make(map[string]struct{}, 32)
	for _, g := range groups {
		for _, a := range ps.groupActions[g] {
			set[a] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for a := range set {
		out = append(out, a)
	}
	sort.Strings(out)
	return out
}

func toPulumiStrings(ss []string) pulumi.StringArray {
	arr := make([]pulumi.StringInput, 0, len(ss))
	for _, s := range ss {
		arr = append(arr, pulumi.String(s))
	}
	return pulumi.StringArray(arr)
}

type StatementSpec struct {
	Groups    []PolicyGroup
	Resources []string // e.g. []string{"*"} or specific ARNs
}

// Helpers for callers:

func AllResources() []string { return []string{"*"} }

func (ps *PolicySet) Statement(resources []string, groups ...PolicyGroup) iam.GetPolicyDocumentStatementArgs {
	return iam.GetPolicyDocumentStatementArgs{
		Actions:   toPulumiStrings(ps.merge(groups...)),
		Resources: toPulumiStrings(resources),
	}
}

func (ps *PolicySet) Build(specs ...StatementSpec) iam.GetPolicyDocumentStatementArray {
	out := make(iam.GetPolicyDocumentStatementArray, 0, len(specs))
	for _, s := range specs {
		out = append(out, ps.Statement(s.Resources, s.Groups...))
	}
	return out
}
