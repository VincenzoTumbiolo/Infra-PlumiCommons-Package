package rest

import (
	policy "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/config/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type RESTModule struct {
	Ctx         *pulumi.Context
	DefaultTags pulumi.StringMap
	Environment pulumi.StringMap

	Policies *policy.PolicySet
}

func New(
	ctx *pulumi.Context,
	defaultTags pulumi.StringMap,
	environment pulumi.StringMap,
) *RESTModule {
	return &RESTModule{
		Ctx:         ctx,
		DefaultTags: defaultTags,
		Environment: environment,
		Policies:    policy.DefaultPolicySet(),
	}
}
