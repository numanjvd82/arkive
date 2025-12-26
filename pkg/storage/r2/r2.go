package r2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"arkive/pkg/SafePtr"
)

type Config struct {
	AccountID       string
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
	region := cfg.Region
	endpoint := cfg.Endpoint
	if endpoint == "" && cfg.AccountID != "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)
	}

	var loadOpts []func(*config.LoadOptions) error
	loadOpts = append(loadOpts, config.WithRegion(region))

	loadOpts = append(loadOpts, config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
	))

	awsCfg, err := config.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &Client{
		s3:      client,
		presign: s3.NewPresignClient(client),
		bucket:  cfg.Bucket,
	}, nil
}

func (c *Client) PresignUpload(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
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

	opts := []func(*s3.PresignOptions){}
	if expires > 0 {
		opts = append(opts, s3.WithPresignExpires(expires))
	}

	out, err := c.presign.PresignPutObject(ctx, input, opts...)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) PresignDownload(ctx context.Context, key string, expires time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("key is required")
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	opts := []func(*s3.PresignOptions){}
	if expires > 0 {
		opts = append(opts, s3.WithPresignExpires(expires))
	}

	out, err := c.presign.PresignGetObject(ctx, input, opts...)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) CreateMultipartUpload(ctx context.Context, key string, contentType string) (string, error) {
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

func (c *Client) PresignUploadPart(ctx context.Context, key string, uploadID string, partNumber int32, expires time.Duration) (string, error) {
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

	opts := []func(*s3.PresignOptions){}
	if expires > 0 {
		opts = append(opts, s3.WithPresignExpires(expires))
	}

	out, err := c.presign.PresignUploadPart(ctx, input, opts...)
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (c *Client) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, parts []types.CompletedPart) error {
	if key == "" {
		return errors.New("key is required")
	}
	if uploadID == "" {
		return errors.New("uploadID is required")
	}
	if len(parts) == 0 {
		return errors.New("parts are required")
	}

	_, err := c.s3.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: parts,
		},
	})
	return err
}

func (c *Client) AbortMultipartUpload(ctx context.Context, key string, uploadID string) error {
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

func (c *Client) ListParts(ctx context.Context, key string, uploadID string) ([]types.Part, error) {
	if key == "" {
		return nil, errors.New("key is required")
	}
	if uploadID == "" {
		return nil, errors.New("uploadID is required")
	}

	input := &s3.ListPartsInput{
		Bucket:   aws.String(c.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	}
	pager := s3.NewListPartsPaginator(c.s3, input)
	parts := make([]types.Part, 0)
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		parts = append(parts, page.Parts...)
	}
	return parts, nil
}

func (c *Client) HeadObjectSize(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, errors.New("key is required")
	}

	out, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, err
	}

	return SafePtr.Int64(out.ContentLength), nil
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
