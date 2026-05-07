package s3client

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"arkive/pkg/header"
	"arkive/pkg/storage"
)

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Bucket          string
	Endpoint        string
	Region          string
	UsePathStyle    bool
}

type Client struct {
	s3      *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	var loadOpts []func(*config.LoadOptions) error
	if cfg.Region != "" {
		loadOpts = append(loadOpts, config.WithRegion(cfg.Region))
	}
	loadOpts = append(loadOpts, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
	))

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &Client{
		s3:      client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
	}, nil
}

func (c *Client) PresignUpload(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("key is required")
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	var opts []func(*s3.PresignOptions)
	if expires > 0 {
		opts = append(opts, s3.WithPresignExpires(expires))
	}

	out, err := c.presign.PresignPutObject(ctx, input, opts...)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) PresignDownload(ctx context.Context, key, filename, disposition string, expires time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("key is required")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}
	contentDisposition := header.BuildContentDisposition(filename, disposition)
	if contentDisposition != "" {
		input.ResponseContentDisposition = aws.String(contentDisposition)
	}

	var opts []func(*s3.PresignOptions)
	if expires > 0 {
		opts = append(opts, s3.WithPresignExpires(expires))
	}

	out, err := c.presign.PresignGetObject(ctx, input, opts...)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("key is required")
	}

	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (c *Client) CreateMultipartUpload(ctx context.Context, key, contentType string) (string, error) {
	if key == "" {
		return "", errors.New("key is required")
	}

	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	out, err := c.s3.CreateMultipartUpload(ctx, input)
	if err != nil {
		return "", err
	}
	return aws.ToString(out.UploadId), nil
}

func (c *Client) PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int32, expires time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("key is required")
	}
	if uploadID == "" {
		return "", errors.New("uploadID is required")
	}
	if partNumber <= 0 {
		return "", errors.New("partNumber must be greater than 0")
	}

	input := &s3.UploadPartInput{
		Bucket:     aws.String(c.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(partNumber),
	}

	var opts []func(*s3.PresignOptions)
	if expires > 0 {
		opts = append(opts, s3.WithPresignExpires(expires))
	}

	out, err := c.presign.PresignUploadPart(ctx, input, opts...)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	if key == "" {
		return errors.New("key is required")
	}
	if uploadID == "" {
		return errors.New("uploadID is required")
	}
	if len(parts) == 0 {
		return errors.New("parts are required")
	}

	completed := make([]types.CompletedPart, 0, len(parts))
	for _, part := range parts {
		if part.PartNumber <= 0 || part.ETag == "" {
			return errors.New("invalid completed part")
		}
		partNumber := part.PartNumber
		etag := part.ETag
		completed = append(completed, types.CompletedPart{
			PartNumber: &partNumber,
			ETag:       &etag,
		})
	}

	_, err := c.s3.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completed,
		},
	})
	return err
}

func (c *Client) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	if key == "" {
		return errors.New("key is required")
	}
	if uploadID == "" {
		return errors.New("uploadID is required")
	}

	_, err := c.s3.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	return err
}
