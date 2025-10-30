package dto

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type ListenerInput struct {
	AwsLbArn       string           // var.aws_lb_arn
	Port           int              // var.lb_forward_listener_port
	Protocol       string           // var.lb_forward_listener_protocol
	CertificateArn *string          // var.lb_certificate_arn (usata solo se Protocol == "HTTPS")
	Tags           pulumi.StringMap // var.tags
}

type LoadBalancerInput struct {
	LbName            string           // var.lb_name
	LbType            string           // var.lb_type ("application" | "network" | "gateway")
	LbSecurityGroupId *pulumi.IDOutput // var.lb_security_group_id (usata solo se LbType == "application")
	LbSubnetIds       []string         // var.lb_subnet_ids
	LogBucket         string           // var.log_bucket
	Tags              pulumi.StringMap // var.tags
}

type TargetGroupInput struct {
	Name       string // var.lb_forward_target_group_name
	Port       int    // var.lb_forward_target_group_port
	TargetType string // var.lb_target_type (es. "instance" | "ip" | "lambda" | "alb")
	Protocol   string // var.protocol (es. "HTTP" | "HTTPS" | "TCP")
	VpcId      string // var.vpc_id

	// health_check {...}
	HealthCheckPath               string // var.health_check_path
	HealthCheckPort               string // var.health_check_port (può essere numero o "traffic-port")
	HealthCheckProtocol           string // var.health_check_protocol
	HealthCheckHealthyThreshold   int    // var.health_check_healthy_threshold
	HealthCheckUnhealthyThreshold int    // var.health_check_unhealthy_threshold
	HealthCheckMatcher            string // var.health_check_matcher (es. "200-399")

	Tags pulumi.StringMap // var.tags
}

type ElbModuleInput struct {
	// Naming
	Env           string // var.env
	ProjectPrefix string // var.tags.project_prefix (usato nei nomi risorse)

	// Network
	PrivateSubnetIds []string // local.private_subnets_id (già calcolati: 2 subnets)
	VpcId            string   // var.vpcId

	// Logs
	LogBucket string // module.lb_logs.s3.bucket

	// ALB
	AlbSecurityGroupId  string  // aws_security_group.alb.id
	AlbListenerPort     int     // var.alb_forward_listener_port
	AlbListenerProtocol string  // var.alb_forward_listener_protocol (es. "HTTPS")
	CertificateArn      *string // var.certificate_arn
	SslPolicy           *string // var.ssl_policy

	// NLB
	NlbListenerPort     int    // var.nlb_forward_listener_port
	NlbListenerProtocol string // var.nlb_forward_listener_protocol

	// TG (per NLB)
	NlbTgProtocol string // var.nlb_forward_target_group_protocol
	NlbTgPort     int    // var.nlb_forward_target_group_port

	// Listener Rule
	NlbListenerRulePriority int // var.nlb_listener_rule_priority

	// Tags base
	Tags pulumi.StringMap // local.tags
}

// ====== Output ======

type ElbModuleResources struct {
	Alb            *lb.LoadBalancer
	AlbListener    *lb.Listener
	AlbHealthRule  *lb.ListenerRule
	Nlb            *lb.LoadBalancer
	NlbTargetGroup *lb.TargetGroup
	NlbListener    *lb.Listener
	TgAttachment   *lb.TargetGroupAttachment
}
