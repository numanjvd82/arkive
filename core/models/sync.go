package models

import "time"

type SyncEntriesResponse struct {
	Entries []SyncEntry `json:"entries"`
}

type SyncEntry struct {
	Type              string     `json:"type"`
	ID                string     `json:"id"`
	FolderID          *string    `json:"folder_id,omitempty"`
	ParentFolderID    *string    `json:"parent_folder_id,omitempty"`
	EncryptedMetadata string     `json:"encrypted_metadata,omitempty"`
	EncryptedFileKey  string     `json:"encrypted_file_key,omitempty"`
	EncryptedManifest string     `json:"encrypted_manifest,omitempty"`
	EncryptedName     string     `json:"encrypted_name,omitempty"`
	UpdatedAt         time.Time  `json:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at"`
	PurgedAt          *time.Time `json:"purged_at,omitempty"`
}
