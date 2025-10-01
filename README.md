## Packages Overview

### infrastructure/commons/vpc

- **Purpose:** Contains utilities for working with AWS VPC resources.
- **Key File:** [`vpc.go`](infrastructure/commons/vpc/vpc.go)
- **Main Function:** [`GetVpcConfigArgs`](infrastructure/commons/vpc/vpc.go) - Generates VPC configuration arguments for Lambda functions.

### infrastructure/core

- **Purpose:** Core utilities for Pulumi resource management.
- **Key File:** [`pulumi.go`](infrastructure/core/pulumi.go)
- **Main Function:** [`MergeStringMap`](infrastructure/core/pulumi.go) - Merges multiple Pulumi string maps, useful for combining resource tags.

### infrastructure/lambda

- **Purpose:** Provides high-level functions to create AWS Lambda resources with different deployment strategies.
- **Key Files:**
  - [`lambda.go`](infrastructure/lambda/lambda.go): Entry point for creating Lambda functions based on type (EFS or S3).
  - [`lambda.efs.go`](infrastructure/lambda/lambda.efs.go): Logic for deploying Lambda functions with EFS.
  - [`lambda.s3.go`](infrastructure/lambda/lambda.s3.go): Logic for deploying Lambda functions from S3.
- **DTOs:** [`dto/lambda.dto.go`](infrastructure/lambda/dto/lambda.dto.go) - Data transfer objects for Lambda configuration.

### infrastructure/lambda/core

- **Purpose:** Lambda-related utilities.
- **Key Files:**
  - [`archive.go`](infrastructure/lambda/core/archive.go): Builds source zip archives for Lambda deployment.
  - [`cloudwatch.go`](infrastructure/lambda/core/cloudwatch.go): Creates CloudWatch alarms for Lambda monitoring.

## Usage

- Use the functions in `infrastructure/lambda` to provision Lambda functions with Pulumi.
- Utilities in `infrastructure/commons/vpc` and `infrastructure/core` help with networking and resource tagging.

## Testing

Run all tests:

```sh
make test.all
```

Run coverage:

```sh
make test.all.cover
```

---

For more details, see individual package documentation and
