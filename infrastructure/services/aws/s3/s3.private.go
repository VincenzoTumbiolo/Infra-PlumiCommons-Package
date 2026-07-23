package s3

import (
	"encoding/json"
	"fmt"

	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateS3Bucket(ctx *pulumi.Context, in dto.S3BucketInput) (*dto.S3BucketResources, error) {
	// Tags fallback
	tags := in.Tags
	if tags == nil {
		tags = pulumi.StringMap{}
	}

	// Bucket
	bkt, err := s3.NewBucket(ctx, in.Name, &s3.BucketArgs{
		Bucket: pulumi.String(in.Name),
		Tags:   tags,
	})
	if err != nil {
		return nil, err
	}

	// Public access block
	pab, err := s3.NewBucketPublicAccessBlock(ctx, fmt.Sprintf("%s-policy", in.Name), &s3.BucketPublicAccessBlockArgs{
		Bucket:                bkt.ID(),
		BlockPublicAcls:       pulumi.Bool(true),
		BlockPublicPolicy:     pulumi.Bool(true),
		IgnorePublicAcls:      pulumi.Bool(true),
		RestrictPublicBuckets: pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	// Versioning (solo se var.versioned == true)
	var ver *s3.BucketVersioningV2
	if in.Versioned {
		ver, err = s3.NewBucketVersioningV2(ctx, fmt.Sprintf("%s-version", in.Name), &s3.BucketVersioningV2Args{
			Bucket: bkt.ID(),
			VersioningConfiguration: &s3.BucketVersioningV2VersioningConfigurationArgs{
				Status: pulumi.String("Enabled"),
			},
		})
		if err != nil {
			return nil, err
		}
	}

	// SSE: AES256
	sse, err := s3.NewBucketServerSideEncryptionConfigurationV2(ctx, fmt.Sprintf("%s-encription", in.Name), &s3.BucketServerSideEncryptionConfigurationV2Args{
		Bucket: bkt.Bucket, // usa il nome del bucket
		Rules: s3.BucketServerSideEncryptionConfigurationV2RuleArray{
			&s3.BucketServerSideEncryptionConfigurationV2RuleArgs{
				ApplyServerSideEncryptionByDefault: &s3.BucketServerSideEncryptionConfigurationV2RuleApplyServerSideEncryptionByDefaultArgs{
					SseAlgorithm: pulumi.String("AES256"),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// CORS
	cors, err := s3.NewBucketCorsConfigurationV2(ctx, fmt.Sprintf("%s-cors", in.Name), &s3.BucketCorsConfigurationV2Args{
		Bucket: bkt.Bucket,
		CorsRules: s3.BucketCorsConfigurationV2CorsRuleArray{
			&s3.BucketCorsConfigurationV2CorsRuleArgs{
				AllowedHeaders: pulumi.ToStringArray([]string{"*"}),
				AllowedMethods: pulumi.ToStringArray([]string{"GET", "HEAD", "PUT", "POST", "DELETE"}),
				AllowedOrigins: pulumi.ToStringArray([]string{"*"}),
				MaxAgeSeconds:  pulumi.Int(3000),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Bucket Policy (da policy_statements)
	var pol *s3.BucketPolicy
	if len(in.PolicyStatements) > 0 {
		policyJSON, err := buildBucketPolicyJSON(in.PolicyStatements)
		if err != nil {
			return nil, err
		}
		pol, err = s3.NewBucketPolicy(ctx, fmt.Sprintf("%s-bucket-policy", in.Name), &s3.BucketPolicyArgs{
			Bucket: bkt.ID(),
			Policy: pulumi.String(policyJSON),
		})
		if err != nil {
			return nil, err
		}
	}

	return &dto.S3BucketResources{
		Bucket:            bkt,
		PublicAccessBlock: pab,
		Versioning:        ver,
		Encryption:        sse,
		Cors:              cors,
		Policy:            pol,
	}, nil
}

// Costruisce il JSON della bucket policy replicando il data "aws_iam_policy_document" dinamico.
func buildBucketPolicyJSON(stmts []dto.S3PolicyStatement) (string, error) {
	type principal struct {
		// es. { "AWS": ["arn:aws:iam::123:role/...", "..."] }
		Any map[string]interface{} `json:"-"`
	}
	type statement struct {
		Sid       string                 `json:"Sid,omitempty"`
		Effect    string                 `json:"Effect"`
		Action    interface{}            `json:"Action"`   // string | []string
		Resource  interface{}            `json:"Resource"` // string | []string
		Principal map[string]interface{} `json:"Principal,omitempty"`
	}
	doc := struct {
		Version   string      `json:"Version"`
		Statement []statement `json:"Statement"`
	}{
		Version: "2012-10-17",
	}

	for _, s := range stmts {
		if s.PrincipalType == "" {
			return "", fmt.Errorf("policy statement %q: PrincipalType mancante", s.Sid)
		}
		if len(s.PrincipalIdentifiers) == 0 {
			return "", fmt.Errorf("policy statement %q: PrincipalIdentifiers mancante", s.Sid)
		}
		stmt := statement{
			Sid:      s.Sid,
			Effect:   s.Effect,
			Action:   normalizeOneOrMany(s.Actions),
			Resource: normalizeOneOrMany(s.Resources),
			Principal: map[string]interface{}{
				s.PrincipalType: normalizeOneOrMany(s.PrincipalIdentifiers),
			},
		}
		doc.Statement = append(doc.Statement, stmt)
	}

	b, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func normalizeOneOrMany(xs []string) interface{} {
	if len(xs) == 1 {
		return xs[0]
	}
	out := make([]string, len(xs))
	copy(out, xs)
	return out
}
