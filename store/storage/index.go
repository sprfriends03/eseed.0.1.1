package storage

import (
	"app/env"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/lifecycle"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	client     *minio.Client
	bucketName string
}

// Required bucket names for the platform
var requiredBuckets = []string{
	"profile-images",
	"documents",
	"plant-images",
	"harvest-images",
	"kyc-documents",
	"nft-metadata",
}

func New() *Storage {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	uri, err := url.Parse(env.MinioUri)
	if err != nil {
		logrus.Fatalln("Minio", err)
	}

	minioHost := uri.Hostname()
	minioPort := uri.Port()
	minioUser := uri.User.Username()
	minioPass, _ := uri.User.Password()
	minioBucket := strings.ReplaceAll(uri.Path, "/", "")

	endpoint := fmt.Sprintf("%v:%v", minioHost, minioPort)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioUser, minioPass, ""),
		Secure: false, // Set to true for https
	})
	if err != nil {
		logrus.Fatalln("Minio", err)
	}

	// Create the main bucket if it doesn't exist
	if ok, err := client.BucketExists(ctx, minioBucket); err != nil {
		logrus.Fatalln("Minio", err)
	} else if !ok {
		err = client.MakeBucket(ctx, minioBucket, minio.MakeBucketOptions{})
		if err != nil {
			logrus.Fatalln("Failed to create main bucket:", err)
		}
		err = client.SetBucketPolicy(ctx, minioBucket, policy(minioBucket))
		if err != nil {
			logrus.Warnln("Failed to set bucket policy:", err)
		}
	}

	// Create required buckets
	for _, bucket := range requiredBuckets {
		bucketName := minioBucket + "-" + bucket
		exists, err := client.BucketExists(ctx, bucketName)
		if err != nil {
			logrus.Warnf("Failed to check if bucket %s exists: %v", bucketName, err)
			continue
		}

		if !exists {
			err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				logrus.Warnf("Failed to create bucket %s: %v", bucketName, err)
				continue
			}

			// Set bucket policy
			err = client.SetBucketPolicy(ctx, bucketName, policy(bucketName))
			if err != nil {
				logrus.Warnf("Failed to set policy for bucket %s: %v", bucketName, err)
			}

			// Configure bucket lifecycle rules
			config := lifecycle.NewConfiguration()
			config.Rules = []lifecycle.Rule{
				{
					ID:     bucketName + "-expiry",
					Status: "Enabled",
					Expiration: lifecycle.Expiration{
						Days: 365, // Set expiration to 1 year for all objects
					},
				},
			}

			err = client.SetBucketLifecycle(ctx, bucketName, config)
			if err != nil {
				logrus.Warnf("Failed to set lifecycle policy for bucket %s: %v", bucketName, err)
			}
		}
	}

	fmt.Printf("Minio connected %v\n", env.MinioUri)

	return &Storage{client, minioBucket}
}

func policy(bucketName string) string {
	return fmt.Sprintf(`
		{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": [
							"*"
						]
					},
					"Action": [
						"s3:GetBucketLocation",
						"s3:ListBucket",
						"s3:ListBucketMultipartUploads"
					],
					"Resource": [
						"arn:aws:s3:::%v"
					]
				},
				{
					"Effect": "Allow",
					"Principal": {
						"AWS": [
							"*"
						]
					},
					"Action": [
						"s3:AbortMultipartUpload",
						"s3:DeleteObject",
						"s3:GetObject",
						"s3:ListMultipartUploadParts",
						"s3:PutObject"
					],
					"Resource": [
						"arn:aws:s3:::%v/*"
					]
				}
			]
		}`,
		bucketName,
		bucketName,
	)
}

func (s Storage) Uri() string {
	return env.MinioUri
}

func (s Storage) Instance() *minio.Client {
	return s.client
}

func (s Storage) Download(ctx context.Context, objectName string) (*minio.Object, error) {
	return s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
}

func (s Storage) Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (minio.UploadInfo, error) {
	return s.client.PutObject(ctx, s.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{CacheControl: "public, max-age=2592000", ContentType: contentType})
}

// GetBucketForType returns the appropriate bucket for a given file type
func (s Storage) GetBucketForType(fileType string) string {
	switch fileType {
	case "profile":
		return s.bucketName + "-profile-images"
	case "document":
		return s.bucketName + "-documents"
	case "plant":
		return s.bucketName + "-plant-images"
	case "harvest":
		return s.bucketName + "-harvest-images"
	case "kyc":
		return s.bucketName + "-kyc-documents"
	case "nft":
		return s.bucketName + "-nft-metadata"
	default:
		return s.bucketName
	}
}

// UploadToTypeBucket uploads a file to the appropriate type-specific bucket
func (s Storage) UploadToTypeBucket(ctx context.Context, fileType, objectName string, reader io.Reader, objectSize int64, contentType string) (minio.UploadInfo, error) {
	bucket := s.GetBucketForType(fileType)
	return s.client.PutObject(ctx, bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		CacheControl: "public, max-age=2592000",
		ContentType:  contentType,
	})
}

// DownloadFromTypeBucket downloads a file from the appropriate type-specific bucket
func (s Storage) DownloadFromTypeBucket(ctx context.Context, fileType, objectName string) (*minio.Object, error) {
	bucket := s.GetBucketForType(fileType)
	return s.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
}

// UploadProfileImage uploads a profile image to the profile-images bucket
func (s Storage) UploadProfileImage(ctx context.Context, reader io.Reader, filename string) (string, error) {
	// Generate unique filename with timestamp
	objectName := fmt.Sprintf("%d-%s", time.Now().Unix(), filename)

	// Get file size by reading into memory (for small profile images this is acceptable)
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	// Create new reader from data
	newReader := strings.NewReader(string(data))
	objectSize := int64(len(data))

	// Determine content type from filename extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(filename), ".jpg") || strings.HasSuffix(strings.ToLower(filename), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(filename), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(filename), ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(strings.ToLower(filename), ".webp") {
		contentType = "image/webp"
	}

	_, err = s.UploadToTypeBucket(ctx, "profile", objectName, newReader, objectSize, contentType)
	if err != nil {
		return "", err
	}

	return objectName, nil
}

// DeleteProfileImage deletes a profile image from the profile-images bucket
func (s Storage) DeleteProfileImage(ctx context.Context, objectName string) error {
	bucket := s.GetBucketForType("profile")
	return s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}
