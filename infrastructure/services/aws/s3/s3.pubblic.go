package s3

import (
	"fmt"

	dto "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/infrastructure/dto/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreatePublicS3Bucket(ctx *pulumi.Context, in dto.PublicS3BucketInput) (*dto.PublicS3BucketResources, error) {
	// Tags fallback
	tags := in.Tags
	if tags == nil {
		tags = pulumi.StringMap{}
	}

	// --- Bucket ---
	bkt, err := s3.NewBucket(ctx, in.Name, &s3.BucketArgs{
		Bucket: pulumi.String(in.Name),
		Tags:   tags,
	})
	if err != nil {
		return nil, err
	}

	// --- Public Access Block (tutto disabilitato = bucket pubblico) ---
	pab, err := s3.NewBucketPublicAccessBlock(ctx, fmt.Sprintf("%s-bucket-policy", in.Name), &s3.BucketPublicAccessBlockArgs{
		Bucket:                bkt.ID(),
		BlockPublicAcls:       pulumi.Bool(false),
		BlockPublicPolicy:     pulumi.Bool(false),
		IgnorePublicAcls:      pulumi.Bool(false),
		RestrictPublicBuckets: pulumi.Bool(false),
	})
	if err != nil {
		return nil, err
	}

	// --- Versioning (solo se richiesto) ---
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

	// --- Server Side Encryption AES256 ---
	sse, err := s3.NewBucketServerSideEncryptionConfigurationV2(ctx, fmt.Sprintf("%s-encription", in.Name), &s3.BucketServerSideEncryptionConfigurationV2Args{
		Bucket: bkt.Bucket,
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

	return &dto.PublicS3BucketResources{
		Bucket:            bkt,
		PublicAccessBlock: pab,
		Versioning:        ver,
		Encryption:        sse,
	}, nil
}
