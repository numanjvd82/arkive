package models

import "time"

type ListEntriesPageInput struct {
	UserID         string
	FolderID       *string
	IncludeDeleted bool
	Cursor         *SyncEntriesCursor
	Limit          int
}

type SyncEntriesResponse struct {
	Entries    []SyncEntry `json:"entries"`
	NextCursor *string     `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
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

type SyncEntriesCursor struct {
	UpdatedAt time.Time
	Type      string
	ID        string
}
