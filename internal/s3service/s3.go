package s3service

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/felixsolom/fetch-duck/internal/config"
)

type Service struct {
	S3Client   *s3.Client
	BucketName string
}

func New(cfg config.AWSConfig) (*Service, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(creds))
	if err != nil {
		return nil, fmt.Errorf("Failed to load AWS config")
	}

	s3Client := s3.NewFromConfig(awsCfg)

	return &Service{
		S3Client:   s3Client,
		BucketName: cfg.BucketName,
	}, nil
}
