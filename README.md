# Infra-PulumiCommons-Package

`Infra-PulumiCommons-Package` is a Go library that streamlines cloud infrastructure management with Pulumi. It offers reusable modules, services, and utilities for AWS resources, including Lambda functions, API Gateway, RDS databases, and networking components.

## Features

- **DTOs**: Define resource configurations with Data Transfer Objects.
- **Mappers**: Map configurations to Pulumi resource arguments.
- **Modules**: Modular components for REST APIs, databases, and Lambda functions.
- **Services**: Core services for deploying and managing AWS resources.

## Folder Structure

### `dto/`

Data Transfer Objects for resource configurations:

- `apigw.dto.go`: API Gateway DTOs
- `db.dto.go`: Database DTOs
- `lambda.dto.go`: Lambda DTOs
- `sg.dto.go`: Security Group DTOs

### `mappers/`

Utilities for mapping configurations to Pulumi resource arguments:

- `pulumi.mappers.go`: General Pulumi mappers
- `vpc.mappers.go`: VPC configuration mappers

### `modules/`

Modular infrastructure components:

- `rest/`: REST API modules
  - `apigw.go`: API Gateway
  - `db.go`: Database
  - `lambda.go`: Lambda
  - `module.go`: Core module logic
  - `config/`: AWS policies and tags

### `services/`

Services for cloud resource management:

- `apigw/`: API Gateway services
  - `core/`: Deployment and resource management
- `lambda/`: Lambda services
  - `lambda.efs.go`: Lambda with EFS
  - `lambda.s3.go`: Lambda with S3
  - `core/`: Lambda utilities
- `network/`: Networking services
  - `sg.go`: Security Group management

## Requirements

- **Go**: v1.24.2 or newer
- **Pulumi**: For infrastructure management

## Installation

Add the package to your Go module:

```bash
go get github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package
```

## Usage

Example: Creating a Security Group

```go
package main

import (
  "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/dto"
  "github.com/VincenzoTumbiolo/Infra-PlumiCommons-Package/services/network"
  "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
  pulumi.Run(func(ctx *pulumi.Context) error {
    sgConfig := dto.SecurityGroupDTO{
      Name:        "example-sg",
      Description: "Example security group",
      Ingress: []dto.SecurityGroupRuleDTO{
        {
          Protocol:  "tcp",
          FromPort:  80,
          ToPort:    80,
          CidrBlock: "0.0.0.0/0",
        },
      },
      Egress: []dto.SecurityGroupRuleDTO{
        {
          Protocol:  "tcp",
          FromPort:  0,
          ToPort:    65535,
          CidrBlock: "0.0.0.0/0",
        },
      },
    }

    _, err := network.CreateSecurityGroup(ctx, sgConfig)
    return err
  })
}
```

## Dependencies

Major dependencies:

```bash
github.com/pulumi/pulumi-aws/sdk/v7
github.com/pulumi/pulumi/sdk/v3
```

## Contributing

Contributions are welcome! Open an issue or submit a pull request to propose changes.

## License

MIT License. See the LICENSE file for details.

## Author

Vincenzo Tumbiolo
