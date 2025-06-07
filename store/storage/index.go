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

// KYC Document Methods

// UploadKYCDocument uploads a KYC document to the kyc-documents bucket
func (s Storage) UploadKYCDocument(ctx context.Context, memberID, docType, fileType string, reader io.Reader, filename string) (string, error) {
	// Generate unique object name: memberID/docType/fileType/timestamp-filename
	objectName := fmt.Sprintf("%s/%s/%s/%d-%s", memberID, docType, fileType, time.Now().Unix(), filename)

	// Get file size by reading into memory
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Validate file size (max 10MB for KYC documents)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if int64(len(data)) > maxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	// Create new reader from data
	newReader := strings.NewReader(string(data))
	objectSize := int64(len(data))

	// Determine content type from filename extension
	contentType := getContentTypeFromFilename(filename)

	// Upload to KYC bucket
	_, err = s.UploadToTypeBucket(ctx, "kyc", objectName, newReader, objectSize, contentType)
	if err != nil {
		return "", fmt.Errorf("failed to upload KYC document: %w", err)
	}

	return objectName, nil
}

// DeleteKYCDocument deletes a KYC document from the kyc-documents bucket
func (s Storage) DeleteKYCDocument(ctx context.Context, objectName string) error {
	bucket := s.GetBucketForType("kyc")
	err := s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete KYC document: %w", err)
	}
	return nil
}

// GetKYCDocumentURL generates a presigned URL for accessing a KYC document
func (s Storage) GetKYCDocumentURL(ctx context.Context, objectName string) (string, error) {
	bucket := s.GetBucketForType("kyc")

	// Generate presigned URL valid for 1 hour
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, objectName, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// ValidateKYCFile validates file type, size, and basic security checks for KYC uploads
func (s Storage) ValidateKYCFile(filename string, fileSize int64, fileContent []byte) error {
	// Check file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024) // 10MB
	if fileSize > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", fileSize, maxSize)
	}

	// Check file extension
	allowedExtensions := []string{".jpg", ".jpeg", ".png", ".pdf", ".tiff", ".tif"}
	isAllowed := false
	lowerFilename := strings.ToLower(filename)
	for _, ext := range allowedExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fmt.Errorf("file type not allowed. Allowed types: %v", allowedExtensions)
	}

	// Basic file content validation (magic number check)
	if len(fileContent) < 10 {
		return fmt.Errorf("file appears to be corrupted or empty")
	}

	// Check magic numbers for common file types
	validMagicNumber := false

	// JPEG magic numbers
	if len(fileContent) >= 3 && fileContent[0] == 0xFF && fileContent[1] == 0xD8 && fileContent[2] == 0xFF {
		validMagicNumber = true
	}
	// PNG magic number
	if len(fileContent) >= 8 && fileContent[0] == 0x89 && fileContent[1] == 0x50 && fileContent[2] == 0x4E && fileContent[3] == 0x47 {
		validMagicNumber = true
	}
	// PDF magic number
	if len(fileContent) >= 4 && string(fileContent[0:4]) == "%PDF" {
		validMagicNumber = true
	}
	// TIFF magic numbers
	if len(fileContent) >= 4 && ((fileContent[0] == 0x49 && fileContent[1] == 0x49 && fileContent[2] == 0x2A && fileContent[3] == 0x00) ||
		(fileContent[0] == 0x4D && fileContent[1] == 0x4D && fileContent[2] == 0x00 && fileContent[3] == 0x2A)) {
		validMagicNumber = true
	}

	if !validMagicNumber {
		return fmt.Errorf("file content does not match expected file type")
	}

	return nil
}

// Helper function to determine content type from filename
func getContentTypeFromFilename(filename string) string {
	lowerFilename := strings.ToLower(filename)

	if strings.HasSuffix(lowerFilename, ".jpg") || strings.HasSuffix(lowerFilename, ".jpeg") {
		return "image/jpeg"
	} else if strings.HasSuffix(lowerFilename, ".png") {
		return "image/png"
	} else if strings.HasSuffix(lowerFilename, ".pdf") {
		return "application/pdf"
	} else if strings.HasSuffix(lowerFilename, ".tiff") || strings.HasSuffix(lowerFilename, ".tif") {
		return "image/tiff"
	} else if strings.HasSuffix(lowerFilename, ".gif") {
		return "image/gif"
	} else if strings.HasSuffix(lowerFilename, ".webp") {
		return "image/webp"
	}

	return "application/octet-stream"
}
