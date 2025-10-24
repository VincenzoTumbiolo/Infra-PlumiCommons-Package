package dto

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

type EnvVar struct {
	Name  string             `json:"name"`
	Value pulumi.StringInput `json:"value"`
}

type PortMapping struct {
	Protocol      string `json:"protocol"`      // "tcp" | "udp"
	ContainerPort int    `json:"containerPort"` // es. 8080
	HostPort      int    `json:"hostPort"`      // per Fargate di solito = containerPort
}

type DDogSidecar struct {
	Enable       bool
	Image        string
	Env          []EnvVar
	PortMappings []PortMapping
}

type ECSInput struct {
	// Naming
	Env           string
	ProjectPrefix string
	ServiceName   string // es. "main"
	Region        string
	AccountID     string
	Tags          pulumi.StringMap

	// Task
	EcrImageUrl   pulumi.StringInput
	TaskCPU       string // "256"
	TaskMemory    string // "512"
	ContainerPort int
	HostPort      int
	EnvVars       []EnvVar

	// Networking
	SubnetIds        []pulumi.StringInput // public subnets
	SecurityGroupIds []pulumi.StringInput
	AssignPublicIp   bool // true per public subnet senza NAT/LB

	// Desired count & Autoscaling
	DesiredCount int
	MinCapacity  int
	MaxCapacity  int
	TargetCPU    float64 // 60
	TargetMemory float64 // 80

	// Logs
	LogGroupName     string // "/aws/ecs/dev-psicoapp-main"
	LogRetentionDays int    // 14

	// IAM extra
	AttachS3FullAccess     bool
	CustomInlinePolicyJSON *string // inline policy JSON sul taskRole

	// Log subscription → Lambda (opzionale)
	LogSubscriptionLambdaName *string
	LogSubscriptionLambdaArn  *string

	// Sidecar opzionale
	DDog DDogSidecar

	// ---- DNS (opzionale, per collegare Route53 al task pubblico) ----
	// Zona già esistente (es. Z123ABC...), e recordName tipo "api" o "api.miodominio.com"
	// Se passi "api", la funzione creerà "api.<zoneName>"
	HostedZoneId    *string // es. "Z123ABC..."
	RecordName      *string // es. "api" oppure FQDN "api.miodominio.com"
	RecordTTL       *int    // default 60
	BackendPublicIp *string // IP pubblico del task, passato dalla pipeline
}

type ECSOutput struct {
	ClusterName          pulumi.StringOutput
	ServiceName          pulumi.StringOutput
	ServiceArn           pulumi.StringOutput
	TaskDefinitionArn    pulumi.StringOutput
	TaskExecutionRoleArn pulumi.StringOutput
	TaskRoleArn          pulumi.StringOutput
	LogGroupName         pulumi.StringOutput
}
