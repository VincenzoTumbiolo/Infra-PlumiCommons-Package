package s3

import (
	"context"
	"strconv"
	"time"

	"io"
	"sort"

	"github.com/amplifon-x/ax-go-application-layer/v5/slicex"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
)

// S3Client is a wrapper for the AWS s3.Client
type S3Client struct {
	*s3.Client
}

// New builds a new S3Client with the provided HTTP client
func New(ctx context.Context, httpClient config.HTTPClient, optFns ...func(*s3.Options)) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	var mod S3Client
	mod.Client = s3.NewFromConfig(cfg, optFns...)

	return &mod, nil
}

// ListBuckets is a proxy for the original method
func (s *S3Client) ListBuckets(ctx context.Context) (*s3.ListBucketsOutput, error) {
	return s.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
}

type FileUploadInfo struct {
	Mimetype string `json:"mimetype"`
	Size     int64  `json:"size"`
}

// UploadFile puts an object in the bucket
func (s *S3Client) UploadFile(ctx context.Context,
	bucketName, fileName string,
	file io.Reader,
	fileUploadInfo FileUploadInfo,
) error {
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(fileName),
		Body:          file,
		ContentType:   aws.String(fileUploadInfo.Mimetype),
		ContentLength: aws.Int64(fileUploadInfo.Size),
	})

	return err
}

// UploadLargeFile puts an object in the bucket via MultiPart uploader
func (s *S3Client) UploadLargeFile(ctx context.Context,
	bucketName, fileName string,
	file io.Reader,
	fileUploadInfo FileUploadInfo,
) error {
	// default part size is 5MB
	_, err := manager.NewUploader(s).Upload(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(fileName),
		Body:          file,
		ContentType:   aws.String(fileUploadInfo.Mimetype),
		ContentLength: aws.Int64(fileUploadInfo.Size),
	})

	return err
}

// Upload a file to S3 that expires after TTL minutes
func (s *S3Client) UploadFileExpiringInMinutes(ctx context.Context,
	bucketName, fileName string,
	file io.Reader,
	fileUploadInfo FileUploadInfo,
	ttlInMinutes int,
) (string, error) {
	expires := time.Now().Add(time.Minute * time.Duration(max(ttlInMinutes, 0)))
	return s.GeneratePUTPresignedURL(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(fileName),
		Body:          file,
		ContentType:   aws.String(fileUploadInfo.Mimetype),
		ContentLength: aws.Int64(fileUploadInfo.Size),
		Expires:       &expires,
	})
}

// Returns a file from a bucket
func (s *S3Client) GetFile(ctx context.Context, bucketName string, key string) (*s3.GetObjectOutput, error) {
	return s.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	})
}

// Copy a file from originFileName to newFileName
func (s *S3Client) CopyFile(ctx context.Context, bucketName, originFileName string, newFileName string) error {
	_, err := s.Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucketName),
		CopySource: aws.String(bucketName + "/" + originFileName),
		Key:        aws.String(newFileName),
	})
	return err
}

// DeleteFile deletes an object from the bucket
func (s *S3Client) DeleteFile(ctx context.Context, bucketName, filename string) error {
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	})

	return err
}

// DeleteFolder deletes all objects with the given path prefix
func (s *S3Client) DeleteFolder(ctx context.Context, bucketName, path string) error {
	objects, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(path),
	})
	if err != nil {
		return err
	}

	if len(objects.Contents) == 0 {
		return nil
	}

	if _, err = s.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{
			Objects: slicex.Map(objects.Contents, func(obj types.Object) types.ObjectIdentifier {
				return types.ObjectIdentifier{Key: obj.Key}
			}),
		},
	}); err != nil {
		return err
	}

	return nil
}

// FetchFileInfo returns the head of an object
func (s *S3Client) FetchFileInfo(ctx context.Context, bucketName, fileName string) (*s3.HeadObjectOutput, error) {
	return s.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	})
}

// GetPresignedURL returns the presigned URL for the specified file version
func (s *S3Client) GetPresignedURL(ctx context.Context, bucketName, folderPath, version string) (string, error) {
	var fileKey string
	if version == "" {
		listObjectInput, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(folderPath),
		})
		if err != nil {
			return "", err
		}

		latestObject := s.FindLatestObject(listObjectInput)
		if latestObject != nil {
			*latestObject.Key = "00001"
		}

		fileKey = *latestObject.Key
	} else {
		fileKey = folderPath + "/" + version
	}

	presignedUrl, err := s.GeneratePresignedURL(ctx, bucketName, fileKey)
	if err != nil {
		return "", err
	}

	return presignedUrl, nil
}

type PresignUploadInput struct {
	Bucket  string
	Path    string
	Version string

	MimeType string
}

// GetPresignedURL returns the upload presigned URL for the specified file version
func (s *S3Client) GetUploadPresignedURL(ctx context.Context, ui *PresignUploadInput) (string, error) {
	var key = ui.Path + "/" + func() string {
		if len(ui.Version) == 0 {
			return strconv.FormatInt(time.Now().UnixNano(), 10)
		}
		return ui.Version
	}()

	presignedUrl, err := s.GeneratePUTPresignedURL(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(ui.Bucket),
		ContentType: aws.String(ui.MimeType),
		Key:         aws.String(key),
	})
	if err != nil {
		return "", err
	}

	return presignedUrl, nil
}

// FindLatestObject returns the latest object in a ListObjectV2 output
func (s *S3Client) FindLatestObject(output *s3.ListObjectsV2Output) *types.Object {
	if len(output.Contents) == 0 {
		return nil
	}

	// Sort the objects by last modified time in descending order
	sort.Slice(output.Contents, func(i, j int) bool {
		return output.Contents[i].LastModified.After(*output.Contents[j].LastModified)
	})

	// Return the first (latest) object
	return &output.Contents[0]
}

// Generate a presinged URL with details
func (s *S3Client) GeneratePUTPresignedURL(ctx context.Context, objectDetails *s3.PutObjectInput) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)
	req, err := presignClient.PresignPutObject(ctx, objectDetails)
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// GeneratePresignedURL generates the URL from the given key
func (s *S3Client) GeneratePresignedURL(ctx context.Context, bucketName string, objectKey string) (string, error) {
	presignClient := s3.NewPresignClient(s.Client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// ListContents returns all objects in a bucket
func (s *S3Client) ListContents(ctx context.Context, bucketName string) ([]types.Object, error) {
	listObjectOutput, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return []types.Object{}, err
	}

	return listObjectOutput.Contents, nil
}

// ListContentsByLastModified returns all objects in a bucket,
// sorted by last modified in descending order
func (s *S3Client) ListContentsByLastModified(ctx context.Context, bucketName string) ([]types.Object, error) {
	listObjectOutput, err := s.Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return []types.Object{}, err
	}

	sort.Slice(listObjectOutput.Contents, func(i, j int) bool {
		return listObjectOutput.Contents[i].LastModified.After(*listObjectOutput.Contents[j].LastModified)
	})

	return listObjectOutput.Contents, nil
}
