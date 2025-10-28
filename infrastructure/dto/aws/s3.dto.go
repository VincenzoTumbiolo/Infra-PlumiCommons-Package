package dto

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ---- DTO ----

type S3PolicyStatement struct {
	Sid                  string   // statement.value.sid
	Actions              []string // statement.value.actions
	Effect               string   // statement.value.effect
	Resources            []string // statement.value.resources
	PrincipalType        string   // statement.value.principal_type       (es. "AWS", "Service")
	PrincipalIdentifiers []string // statement.value.principal_identifier (ARNs, "*" ecc.)
}

type S3BucketInput struct {
	Name             string              // var.s3_name
	Tags             pulumi.StringMap    // var.tags
	Versioned        bool                // var.versioned
	PolicyStatements []S3PolicyStatement // var.policy_statements
}

// ---- Funzione ----

type S3BucketResources struct {
	Bucket            *s3.Bucket
	PublicAccessBlock *s3.BucketPublicAccessBlock
	Versioning        *s3.BucketVersioningV2
	Encryption        *s3.BucketServerSideEncryptionConfigurationV2
	Cors              *s3.BucketCorsConfigurationV2
	Policy            *s3.BucketPolicy
}

type PublicS3BucketInput struct {
	Name      string           // var.s3_name
	Tags      pulumi.StringMap // var.tags
	Versioned bool             // var.versioned
}

type PublicS3BucketResources struct {
	Bucket            *s3.Bucket
	PublicAccessBlock *s3.BucketPublicAccessBlock
	Versioning        *s3.BucketVersioningV2
	Encryption        *s3.BucketServerSideEncryptionConfigurationV2
}
