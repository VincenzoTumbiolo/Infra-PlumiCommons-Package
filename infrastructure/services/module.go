package rest

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	policy "github.com/vtechstudio/infra-pulumicommons-package/infrastructure/config/aws"
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
