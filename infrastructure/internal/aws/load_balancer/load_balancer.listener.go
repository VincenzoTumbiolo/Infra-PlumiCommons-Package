package load_balancer

import (
	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateListener(ctx *pulumi.Context, in dto.ListenerInput) (*lb.Listener, error) {
	// Certificate only if HTTPS (come in Terraform)
	var cert pulumi.StringPtrInput
	if in.Protocol == "HTTPS" && in.CertificateArn != nil {
		cert = pulumi.StringPtr(*in.CertificateArn)
	}

	listener, err := lb.NewListener(ctx, "this", &lb.ListenerArgs{
		LoadBalancerArn: pulumi.String(in.AwsLbArn),
		Port:            pulumi.Int(in.Port),
		Protocol:        pulumi.String(in.Protocol),
		CertificateArn:  cert,
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type: pulumi.String("fixed-response"),
				FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
					ContentType: pulumi.String("text/plain"),
					MessageBody: pulumi.StringPtr("FORWARD ERROR"),
					StatusCode:  pulumi.String("400"),
				},
			},
		},
		Tags: in.Tags,
	})
	if err != nil {
		return nil, err
	}

	return listener, nil
}
