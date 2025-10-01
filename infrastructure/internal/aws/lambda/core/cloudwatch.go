package lambda_core

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/vtechstudio/infra-pulumicommons-package/infrastructure/mappers"
)

func CreateLambdaAlarms(ctx *pulumi.Context, functionName string, tags pulumi.StringMap) error {
	alarmTags := pulumi.StringMap{
		"Name": pulumi.String(fmt.Sprintf("%s-alarm", functionName)),
	}
	mergedTags := mappers.MergeStringMap(tags, alarmTags)

	_, err := cloudwatch.NewMetricAlarm(ctx, fmt.Sprintf("%s-throttles", functionName), &cloudwatch.MetricAlarmArgs{
		Name:               pulumi.String(fmt.Sprintf("%s-alarm-throttles", functionName)),
		ComparisonOperator: pulumi.String("GreaterThanThreshold"),
		EvaluationPeriods:  pulumi.Int(1),
		MetricName:         pulumi.String("Throttles"),
		Namespace:          pulumi.String("AWS/Lambda"),
		Period:             pulumi.Int(60),
		Statistic:          pulumi.String("Sum"),
		Threshold:          pulumi.Float64(0),
		AlarmDescription:   pulumi.String(fmt.Sprintf("This metric monitors %s", functionName)),
		Dimensions: pulumi.StringMap{
			"FunctionName": pulumi.String(functionName),
		},
		Tags: mergedTags,
	})
	if err != nil {
		return err
	}

	_, err = cloudwatch.NewMetricAlarm(ctx, fmt.Sprintf("%s-errors", functionName), &cloudwatch.MetricAlarmArgs{
		Name:               pulumi.String(fmt.Sprintf("%s-alarm-errors", functionName)),
		ComparisonOperator: pulumi.String("GreaterThanThreshold"),
		EvaluationPeriods:  pulumi.Int(1),
		MetricName:         pulumi.String("Errors"),
		Namespace:          pulumi.String("AWS/Lambda"),
		Period:             pulumi.Int(60),
		Statistic:          pulumi.String("Sum"),
		Threshold:          pulumi.Float64(10),
		AlarmDescription:   pulumi.String(fmt.Sprintf("This metric monitors %s", functionName)),
		Dimensions: pulumi.StringMap{
			"FunctionName": pulumi.String(functionName),
		},
		Tags: mergedTags,
	})
	return err
}
