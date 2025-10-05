//go:build e2e
// +build e2e

package s3migrator_test

import (
	"context"
	"embed"
	"fmt"
	"testing"

	"github.com/amplifon-x/ax-go-application-layer/v5/db/s3/s3migrator"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/testcontainers/testcontainers-go"
)

//go:embed testdata/e2e-migrations/*.json
var migrations embed.FS

//go:embed testdata/e2e-assets/*
var assets embed.FS

func TestMigrations(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "adobe/s3mock",
		ExposedPorts: []string{"9090/tcp"},
	}
	s3C, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Could not start s3: %s", err)
	}

	t.Cleanup(func() {
		if err := s3C.Terminate(ctx); err != nil {
			t.Fatalf("Could not stop s3: %s", err)
		}
	})

	endpoint, err := s3C.Endpoint(ctx, "")
	if err != nil {
		t.Error(err)
	}
	testEndpoint = endpoint

	d, err := iofs.New(migrations, "testdata/e2e-migrations")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(testEndpointResolver{}),
	)
	if err != nil {
		t.Fatalf("error loading AWS config: %v", err)
	}
	client := s3.NewFromConfig(cfg)

	bucketName := "test-bucket"
	s3Driver, err := s3migrator.WithClient(context.Background(), client, s3migrator.S3DriverConfig{
		BucketName: bucketName,
		Assets:     assets,
	})
	if err != nil {
		t.Fatalf("error creating S3 driver: %v", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", d, bucketName, s3Driver)
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}

	if err := migrator.Up(); err != nil {
		t.Fatalf("error migrating up: %v", err)
	}

	if err := migrator.Down(); err != nil {
		t.Fatalf("error migrating down: %v", err)
	}
}

var testEndpoint string

type testEndpointResolver struct{}

func (testEndpointResolver) ResolveEndpoint(service, region string, options ...interface{}) (aws.Endpoint, error) {
	return aws.Endpoint{
		URL:               fmt.Sprintf("http://%s", testEndpoint),
		HostnameImmutable: true,
	}, nil
}
