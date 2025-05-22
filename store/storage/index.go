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
	"github.com/sirupsen/logrus"
)

type Storage struct {
	client     *minio.Client
	bucketName string
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

	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(minioUser, minioPass, "")})
	if err != nil {
		logrus.Fatalln("Minio", err)
	}

	if ok, err := client.BucketExists(ctx, minioBucket); err != nil {
		logrus.Fatalln("Minio", err)
	} else if !ok {
		client.MakeBucket(ctx, minioBucket, minio.MakeBucketOptions{})
		client.SetBucketPolicy(ctx, minioBucket, policy(minioBucket))
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
