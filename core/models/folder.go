package models

import "time"

type Folder struct {
	ID                string
	UserID            string
	VaultID           string
	ParentFolderID    *string
	EncryptedName     []byte
	EncryptedMetadata []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	SearchScore       int64
}
