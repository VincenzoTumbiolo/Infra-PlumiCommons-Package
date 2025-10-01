package mappers

import (
	"strings"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	dto "github.com/vtechstudio/infra-pulumicommons-package/infrastructure/dto/aws"
)

func MergeStringMap(tags ...pulumi.StringMap) pulumi.StringMap {
	mergedTags := pulumi.StringMap{}
	for _, val := range tags {
		for k, v := range val {
			mergedTags[k] = v
		}
	}
	return mergedTags
}

func ToPulumiStringMap(input any) pulumi.StringMap {
	if input == nil {
		return pulumi.StringMap{}
	}

	result := pulumi.StringMap{}
	if m, ok := input.(map[string]string); ok {
		for k, v := range m {
			result[k] = pulumi.String(v)
		}
	} else if m, ok := input.(map[string]interface{}); ok {
		for k, v := range m {
			if str, ok := v.(string); ok {
				result[k] = pulumi.String(str)
			}
		}
	}

	return result
}

func MapRulesToIngress(rules []dto.SecurityGroupRule) ec2.SecurityGroupIngressArray {
	var result ec2.SecurityGroupIngressArray
	for _, rule := range rules {
		result = append(result, ec2.SecurityGroupIngressArgs{
			Protocol:    pulumi.String(rule.Protocol),
			FromPort:    pulumi.Int(rule.FromPort),
			ToPort:      pulumi.Int(rule.ToPort),
			CidrBlocks:  pulumi.StringArray{pulumi.String(rule.CidrBlock)},
			Description: pulumi.String(rule.Description),
		})
	}
	return result
}

func MapRulesToEgress(rules []dto.SecurityGroupRule) ec2.SecurityGroupEgressArray {
	var result ec2.SecurityGroupEgressArray
	for _, rule := range rules {
		result = append(result, ec2.SecurityGroupEgressArgs{
			Protocol:    pulumi.String(rule.Protocol),
			FromPort:    pulumi.Int(rule.FromPort),
			ToPort:      pulumi.Int(rule.ToPort),
			CidrBlocks:  pulumi.StringArray{pulumi.String(rule.CidrBlock)},
			Description: pulumi.String(rule.Description),
		})
	}
	return result
}

func NormalizeString(s string) string {
	if s == "" {
		return ""
	}
	s = strings.ToLower(s)
	return strings.ToUpper(string(s[0])) + s[1:]
}
