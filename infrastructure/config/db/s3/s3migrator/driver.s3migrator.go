package s3migrator

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amplifon-x/ax-go-application-layer/v5/opt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/golang-migrate/migrate/v4/database"
)

var _ctx = context.TODO()

const (
	DefaultMigrationsFile = "migrationState.json"
	DriverName            = "s3migrator"
)

func init() {
	driver := S3Driver{}
	database.Register(DriverName, &driver)
}

type S3DriverConfig struct {
	BucketName string
	Assets     embed.FS
}

type S3Driver struct {
	client *s3.Client
	config S3DriverConfig
}

func BucketNameToDSN(bucketName string) string {
	return fmt.Sprintf("%s://%s", DriverName, bucketName)
}

func ContentTypeFromExtension(ext string) string {
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		switch ext {
		case ".mp4":
			contentType = "video/mp4"

		case ".srt":
			contentType = "text/plain"

		case ".zip":
			contentType = "application/zip"
		}
	}

	return contentType
}

func WithClient(ctx context.Context, client *s3.Client, driverConfig S3DriverConfig) (*S3Driver, error) {
	_ctx = ctx

	buckets, err := client.ListBuckets(_ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("error listing buckets: %w", err)
	}

	bucketAlreadyPresent := slices.ContainsFunc(buckets.Buckets, func(b types.Bucket) bool {
		return strings.Compare(opt.Coalesce(b.Name, ""), driverConfig.BucketName) == 0
	})
	if !bucketAlreadyPresent {
		if _, err := client.CreateBucket(_ctx, &s3.CreateBucketInput{
			Bucket: aws.String(driverConfig.BucketName),
		}); err != nil {
			return nil, fmt.Errorf("error creating bucket: %w", err)
		}
	}

	driver := &S3Driver{client: client, config: driverConfig}
	if err := driver.ensureMigrationStateFile(); err != nil {
		return nil, fmt.Errorf("error ensuring migration state file: %w", err)
	}

	return driver, nil
}

func New(ctx context.Context, httpClient config.HTTPClient, driverConfig S3DriverConfig) (*S3Driver, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	return WithClient(ctx, s3.NewFromConfig(cfg), driverConfig)
}

func (driver *S3Driver) ensureMigrationStateFile() error {
	_, err := driver.client.HeadObject(_ctx, &s3.HeadObjectInput{
		Bucket: aws.String(driver.config.BucketName),
		Key:    aws.String(DefaultMigrationsFile),
	})
	if err != nil {
		state := MigrationState{Version: -1}
		jsonState, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("error marshaling initial migration state: %w", err)
		}

		_, err = driver.client.PutObject(_ctx, &s3.PutObjectInput{
			Bucket: aws.String(driver.config.BucketName),
			Key:    aws.String(DefaultMigrationsFile),
			Body:   bytes.NewReader(jsonState),
		})
		if err != nil {
			return fmt.Errorf("error ensuring migration state file: %w", err)
		}
	}

	return nil
}

func (driver *S3Driver) Open(dsn string) (database.Driver, error) {
	bucketName := strings.ReplaceAll(dsn, DriverName+"://", "")
	driverCfg := S3DriverConfig{
		BucketName: bucketName,
	}

	var err error
	driver, err = New(_ctx, nil, driverCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating S3 driver: %w", err)
	}

	return driver, nil
}

func (driver *S3Driver) Close() error {
	return nil
}

func (driver *S3Driver) Lock() error {
	return nil
}

func (driver *S3Driver) Unlock() error {
	return nil
}

func (driver *S3Driver) Run(migration io.Reader) error {
	jsonMigration, err := io.ReadAll(migration)
	if err != nil {
		return fmt.Errorf("error reading migration: %w", err)
	}

	stmts, err := UnmarshalMigrationJSON(jsonMigration)
	if err != nil {
		return fmt.Errorf("error unmarshaling migration JSON: %w", err)
	}

	for _, stmt := range stmts {
		switch stmt := stmt.(type) {
		case UploadStmt:
			slog.InfoContext(_ctx, "Uploading file",
				"filename", stmt.Filename,
				"path", stmt.Path,
			)

			file, err := driver.config.Assets.Open(stmt.Path)
			if err != nil {
				return fmt.Errorf("error opening file in path %s: %w", stmt.Path, err)
			}
			defer file.Close()

			ext := filepath.Ext(stmt.Path)
			contentType := ContentTypeFromExtension(ext)
			params := s3.PutObjectInput{
				ContentType: aws.String(contentType),
				Bucket:      aws.String(driver.config.BucketName),
				Key:         aws.String(stmt.Filename),
				Body:        file,
				Metadata: map[string]string{
					"Content-Type": contentType,
					"Extension":    ext,
				},
			}

			if _, err = driver.client.PutObject(_ctx, &params); err != nil {
				return fmt.Errorf("error uploading file %s: %w", stmt.Filename, err)
			}

		case DeleteStmt:
			slog.InfoContext(_ctx, "Deleting file",
				"filename", stmt.Filename,
			)

			if _, err = driver.client.DeleteObject(_ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(driver.config.BucketName),
				Key:    aws.String(stmt.Filename),
			}); err != nil {
				return fmt.Errorf("error deleting file %s: %w", stmt.Filename, err)
			}

		default:
			return fmt.Errorf("unknown migration action: %T", stmt)
		}
	}

	return nil
}

func (driver *S3Driver) SetVersion(version int, dirty bool) error {
	migrationState := MigrationState{
		Version: version,
		Dirty:   dirty,
	}

	jsonState, err := json.Marshal(migrationState)
	if err != nil {
		return fmt.Errorf("error marshaling migration state: %w", err)
	}

	if _, err := driver.client.PutObject(_ctx, &s3.PutObjectInput{
		Bucket:      aws.String(driver.config.BucketName),
		Key:         aws.String(DefaultMigrationsFile),
		Body:        bytes.NewReader(jsonState),
		ContentType: aws.String(mime.TypeByExtension(".json")),
	}); err != nil {
		return fmt.Errorf("error setting migration state: %w", err)
	}

	return nil
}

func (driver *S3Driver) Version() (version int, dirty bool, err error) {
	stateObj, err := driver.client.GetObject(_ctx, &s3.GetObjectInput{
		Bucket: aws.String(driver.config.BucketName),
		Key:    aws.String(DefaultMigrationsFile),
	})
	if err != nil {
		return 0, false, fmt.Errorf("error getting migration state: %w", err)
	}
	defer stateObj.Body.Close()

	state := MigrationState{}
	if err = json.NewDecoder(stateObj.Body).Decode(&state); err != nil {
		return 0, false, fmt.Errorf("error decoding migration state: %w", err)
	}

	return state.Version, state.Dirty, nil
}

func (driver *S3Driver) Drop() error {
	// no lol
	return nil
}
