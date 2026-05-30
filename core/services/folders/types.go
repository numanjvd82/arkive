package folders

import "arkive/core/models"

type EntryType string

const (
	EntryTypeFolder EntryType = "folder"
	EntryTypeFile   EntryType = "file"
)

type ListEntriesResult struct {
	FolderID     *string
	Page         int
	PageSize     int
	TotalEntries int
	TotalPages   int
	Path         []models.Folder
	Folders      []models.Folder
	Files        []models.File
}

type CreateFolderInput struct {
	UserID            string
	VaultID           string
	ParentFolderID    *string
	EncryptedName     []byte
	EncryptedMetadata []byte
	SearchTokens      []models.FileSearchToken
}

type ListEntriesInput struct {
	UserID   string
	FolderID *string
	Page     int
	PageSize int
}

type MoveEntriesInput struct {
	UserID         string
	TargetFolderID *string
	FileIDs        []string
	FolderIDs      []string
}

type RenameFolderInput struct {
	UserID            string
	FolderID          string
	EncryptedName     []byte
	EncryptedMetadata []byte
	SearchTokens      []models.FileSearchToken
}

type ResolveDeleteScopeInput struct {
	UserID    string
	FileIDs   []string
	FolderIDs []string
}

type DeleteScope struct {
	FolderIDs []string
	FileIDs   []string
}

type DeleteEntriesInput struct {
	UserID    string
	FileIDs   []string
	FolderIDs []string
}

type DeleteEntriesResult struct {
	DeletedFiles   int
	DeletedFolders int
}
