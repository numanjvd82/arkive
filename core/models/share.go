package models

import "time"

type Share struct {
	ID                string
	FileID            string
	OwnerUserID       string
	Token             string
	EncryptedShareKey []byte
	AllowPreview      bool
	AllowDownload     bool
	BurnAfterRead     bool
	PasswordHash      *string
	ExpiresAt         *time.Time
	Status            string
	RevokedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ShareWithFile struct {
	Share
	FileName        string
	FileContentType string
	FileSizeBytes   int64
	FileUpdatedAt   time.Time
}

type ShareLink struct {
	ID                   string
	OwnerUserID          string
	Token                string
	Slug                 *string
	Status               string
	TitleEncrypted       []byte
	DescriptionEncrypted []byte
	EncryptedShareKey    []byte
	CryptoVersion        int16
	PasswordHash         *string
	PasswordMode         string
	ExpiresAt            *time.Time
	RevokedAt            *time.Time
	AllowPreview         bool
	AllowDownload        bool
	CommentsEnabled      bool
	ReactionsEnabled     bool
	BurnAfterRead        bool
	ShowEXIF             bool
	ShowLocation         bool
	StripEXIFDownload    bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ShareItem struct {
	ID           string
	ShareLinkID  string
	ItemType     string
	FileID       string
	DisplayOrder int
	CreatedAt    time.Time
}

type ShareSnapshotFile struct {
	ID                       string
	ShareItemID              string
	FileID                   string
	EncryptedRelativePath    []byte
	EncryptedFileKeyForShare []byte
	DisplayOrder             int
	CreatedAt                time.Time
}

type PublicShareRecord struct {
	ShareID                  string
	Token                    string
	FileID                   string
	VaultID                  string
	EncryptedFileKeyForShare []byte
	EncryptedMetadata        []byte
	EncryptedManifest        []byte
	EncryptedHash            []byte
	EncryptionVersion        int16
	ChunkSize                int64
	TotalChunks              int
	PlaintextSize            int64
	SourceURL                string
}
