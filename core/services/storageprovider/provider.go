package storageprovider

import (
	"context"
	"errors"
	"strings"
	"time"

	"arkive/core/database"
	"arkive/core/models"
	settingsrepo "arkive/core/repositories/settings"
	"arkive/pkg/storage"
	"arkive/pkg/storage/localclient"
	"arkive/pkg/storage/s3client"
)

var ErrStorageNotConfigured = errors.New("storage not configured")

type Provider struct {
	db       database.PgPool
	settings *settingsrepo.Repository
	local    *localclient.Client
}

func New(db database.PgPool, settingsRepo *settingsrepo.Repository, local *localclient.Client) *Provider {
	return &Provider{
		db:       db,
		settings: settingsRepo,
		local:    local,
	}
}

func (p *Provider) PresignUpload(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	settings, err := p.load(ctx)
	if err != nil {
		return "", err
	}
	if settings.Provider == "local" {
		return p.local.PresignUpload(ctx, key, contentType, expires)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return "", err
	}
	return client.PresignUpload(ctx, key, contentType, expires)
}

func (p *Provider) PresignDownload(ctx context.Context, key, filename, disposition string, expires time.Duration) (string, error) {
	settings, err := p.load(ctx)
	if err != nil {
		return "", err
	}
	if settings.Provider == "local" {
		return p.local.PresignDownload(ctx, key, filename, disposition, expires)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return "", err
	}
	return client.PresignDownload(ctx, key, filename, disposition, expires)
}

func (p *Provider) DeleteObject(ctx context.Context, key string) error {
	settings, err := p.load(ctx)
	if err != nil {
		return err
	}
	if settings.Provider == "local" {
		return p.local.DeleteObject(ctx, key)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return err
	}
	return client.DeleteObject(ctx, key)
}

func (p *Provider) ObjectSize(ctx context.Context, key string) (int64, error) {
	settings, err := p.load(ctx)
	if err != nil {
		return 0, err
	}
	if settings.Provider == "local" {
		return p.local.ObjectSize(ctx, key)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return 0, err
	}
	return client.ObjectSize(ctx, key)
}

func (p *Provider) CreateMultipartUpload(ctx context.Context, key, contentType string) (string, error) {
	settings, err := p.load(ctx)
	if err != nil {
		return "", err
	}
	if settings.Provider == "local" {
		return p.local.CreateMultipartUpload(ctx, key, contentType)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return "", err
	}
	return client.CreateMultipartUpload(ctx, key, contentType)
}

func (p *Provider) PresignUploadPart(ctx context.Context, key, uploadID string, partNumber int32, expires time.Duration) (string, error) {
	settings, err := p.load(ctx)
	if err != nil {
		return "", err
	}
	if settings.Provider == "local" {
		return p.local.PresignUploadPart(ctx, key, uploadID, partNumber, expires)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return "", err
	}
	return client.PresignUploadPart(ctx, key, uploadID, partNumber, expires)
}

func (p *Provider) CompleteMultipartUpload(ctx context.Context, key, uploadID string, parts []storage.CompletedPart) error {
	settings, err := p.load(ctx)
	if err != nil {
		return err
	}
	if settings.Provider == "local" {
		return p.local.CompleteMultipartUpload(ctx, key, uploadID, parts)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return err
	}
	return client.CompleteMultipartUpload(ctx, key, uploadID, parts)
}

func (p *Provider) AbortMultipartUpload(ctx context.Context, key, uploadID string) error {
	settings, err := p.load(ctx)
	if err != nil {
		return err
	}
	if settings.Provider == "local" {
		return p.local.AbortMultipartUpload(ctx, key, uploadID)
	}
	client, err := p.s3(ctx, settings)
	if err != nil {
		return err
	}
	return client.AbortMultipartUpload(ctx, key, uploadID)
}

func (p *Provider) ActiveProvider(ctx context.Context) (string, error) {
	settings, err := p.load(ctx)
	if err != nil {
		return "", err
	}
	return settings.Provider, nil
}

func (p *Provider) load(ctx context.Context) (models.StorageSettings, error) {
	settings, err := p.settings.GetStorageSettings(ctx, p.db)
	if err != nil {
		return models.StorageSettings{}, err
	}
	settings.Provider = strings.ToLower(strings.TrimSpace(settings.Provider))
	if settings.Provider != "local" && settings.Provider != "s3" {
		return models.StorageSettings{}, ErrStorageNotConfigured
	}
	return settings, nil
}

func (p *Provider) s3(ctx context.Context, settings models.StorageSettings) (*s3client.Client, error) {
	if settings.S3Region == "" {
		settings.S3Region = "auto"
	}
	return s3client.New(ctx, s3client.Config{
		AccessKeyID:     settings.S3AccessKeyID,
		SecretAccessKey: settings.S3SecretAccessKey,
		SessionToken:    settings.S3SessionToken,
		Bucket:          settings.S3Bucket,
		Endpoint:        settings.S3Endpoint,
		Region:          settings.S3Region,
		UsePathStyle:    settings.S3UsePathStyle,
	})
}
