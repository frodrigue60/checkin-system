package services

import (
	"attendance-api/internal/config"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService interface {
	UploadFile(ctx context.Context, file io.Reader, fileName string, contentType string) (string, error)
	// Returns: uploadURL, publicURL, key, error
	GetPresignedUploadURL(ctx context.Context, userID int, index int, originalName string) (string, string, string, error)
	GetAvatarUploadURL(ctx context.Context, userID int, originalName string) (string, string, string, error)
	GetPublicURL(key string) string
}

type R2StorageService struct {
	client     *s3.Client
	bucketName string
	pubURL     string
}

func NewR2StorageService(cfg *config.Config) (*R2StorageService, error) {
	if cfg.R2AccountID == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" {
		return nil, fmt.Errorf("missing R2 credentials in configuration")
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID),
		}, nil
	})

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithEndpointResolverWithOptions(r2Resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2AccessKeyID, cfg.R2SecretAccessKey, "")),
		awsconfig.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &R2StorageService{
		client:     client,
		bucketName: cfg.R2BucketName,
		pubURL:     cfg.R2PublicURL,
	}, nil
}

func (s *R2StorageService) GetPresignedUploadURL(ctx context.Context, userID int, index int, fileName string) (string, string, string, error) {
	ext := "jpg"
	if strings.Contains(fileName, ".") {
		parts := strings.Split(fileName, ".")
		ext = parts[len(parts)-1]
	}

	key := fmt.Sprintf("attendance/emp_%d_%d_%d.%s", userID, time.Now().Unix(), index, ext)

	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String("image/" + ext),
	})
	if err != nil {
		return "", "", "", err
	}

	publicURL := fmt.Sprintf("%s/%s", s.pubURL, key)
	return req.URL, publicURL, key, nil
}

func (s *R2StorageService) GetAvatarUploadURL(ctx context.Context, userID int, fileName string) (string, string, string, error) {
	ext := "jpg"
	if strings.Contains(fileName, ".") {
		parts := strings.Split(fileName, ".")
		ext = parts[len(parts)-1]
	}

	key := fmt.Sprintf("avatars/user_%d_%d.%s", userID, time.Now().Unix(), ext)

	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String("image/" + ext),
	})
	if err != nil {
		return "", "", "", err
	}

	publicURL := fmt.Sprintf("%s/%s", s.pubURL, key)
	return req.URL, publicURL, key, nil
}

func (s *R2StorageService) UploadFile(ctx context.Context, file io.Reader, fileName string, contentType string) (string, error) {
	key := "reports/" + fileName
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", err
	}

	return key, nil
}

func (s *R2StorageService) GetPublicURL(key string) string {
	if key == "" {
		return ""
	}
	if strings.HasPrefix(key, "http") {
		return key
	}
	return fmt.Sprintf("%s/%s", s.pubURL, key)
}
