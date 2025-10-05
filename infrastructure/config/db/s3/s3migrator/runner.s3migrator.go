package s3migrator

import (
	"context"
	"embed"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
)

func MigrateNoClient(migrationDriver source.Driver, bucketName string, assets embed.FS) error {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return fmt.Errorf("error loading AWS config: %w", err)
	}

	return Migrate(s3.NewFromConfig(cfg), migrationDriver, bucketName, assets)
}

func Migrate(client *s3.Client, migrationDriver source.Driver, bucketName string, assets embed.FS) error {
	s3Driver, err := WithClient(context.Background(), client, S3DriverConfig{
		BucketName: bucketName,
		Assets:     assets,
	})
	if err != nil {
		return err
	}

	migrator, err := migrate.NewWithInstance("iofs", migrationDriver, bucketName, s3Driver)
	if err != nil {
		return err
	}

	if err = migrator.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("Couldn't run migrations", "err", err.Error())
		return err
	}

	return nil
}
