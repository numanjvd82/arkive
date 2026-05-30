package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	filessvc "arkive/core/services/files"
	folderssvc "arkive/core/services/folders"
	"arkive/pkg/apierror"
	"arkive/pkg/errs"
	"arkive/pkg/pagination"
)

type createFolderRequest struct {
	ParentFolderID    *string `json:"parentFolderId"`
	EncryptedName     string  `json:"encryptedName"`
	EncryptedMetadata string  `json:"encryptedMetadata"`
}

type moveEntriesRequest struct {
	TargetFolderID *string  `json:"targetFolderId"`
	FileIDs        []string `json:"fileIds"`
	FolderIDs      []string `json:"folderIds"`
}

type deleteEntriesRequest struct {
	FileIDs   []string `json:"fileIds"`
	FolderIDs []string `json:"folderIds"`
}

type renameEntryRequest struct {
	Type              string               `json:"type"`
	ID                string               `json:"id"`
	EncryptedName     string               `json:"encryptedName"`
	EncryptedMetadata string               `json:"encryptedMetadata"`
	SearchTokens      []searchTokenRequest `json:"searchTokens"`
}

type createFolderResponse struct {
	ID                string  `json:"id"`
	ParentFolderID    *string `json:"parentFolderId"`
	EncryptedName     string  `json:"encryptedName"`
	EncryptedMetadata string  `json:"encryptedMetadata,omitempty"`
}

type deleteEntriesResponse struct {
	DeletedFiles   int `json:"deletedFiles"`
	DeletedFolders int `json:"deletedFolders"`
}

func APICreateFolder(folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		var req createFolderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		encryptedName, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedName))
		if err != nil || len(encryptedName) == 0 {
			apierror.InvalidPayload(c)
			return
		}

		var encryptedMetadata []byte
		if strings.TrimSpace(req.EncryptedMetadata) != "" {
			encryptedMetadata, err = base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedMetadata))
			if err != nil {
				apierror.InvalidPayload(c)
				return
			}
		}

		folder, err := folderService.CreateFolder(c.Request.Context(), folderssvc.CreateFolderInput{
			UserID:            userID.(string),
			VaultID:           userID.(string),
			ParentFolderID:    req.ParentFolderID,
			EncryptedName:     encryptedName,
			EncryptedMetadata: encryptedMetadata,
		})
		if err != nil {
			switch err {
			case folderssvc.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case folderssvc.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "parent_folder_not_found", "Parent folder not found", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Create folder failed")
			}
			return
		}

		resp := createFolderResponse{
			ID:             folder.ID,
			ParentFolderID: folder.ParentFolderID,
			EncryptedName:  base64.StdEncoding.EncodeToString(folder.EncryptedName),
		}
		if len(folder.EncryptedMetadata) > 0 {
			resp.EncryptedMetadata = base64.StdEncoding.EncodeToString(folder.EncryptedMetadata)
		}
		c.JSON(http.StatusCreated, gin.H{"folder": resp})
	}
}

func APIListRootFolderEntries(folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiListFolderEntries(c, folderService, nil)
	}
}

func APIListFolderEntries(folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		folderID := strings.TrimSpace(c.Param("id"))
		if folderID == "" {
			apierror.Write(c, http.StatusNotFound, "folder_not_found", "Folder not found", nil)
			return
		}
		apiListFolderEntries(c, folderService, &folderID)
	}
}

func apiListFolderEntries(c *gin.Context, folderService *folderssvc.Service, folderID *string) {
	userID, ok := c.Get("user_id")
	if !ok {
		apierror.Unauthorized(c)
		return
	}

	result, err := folderService.ListEntries(c.Request.Context(), folderssvc.ListEntriesInput{
		UserID:   userID.(string),
		FolderID: folderID,
		Page:     pagination.ParsePageParam(c.DefaultQuery("page", "1")),
		PageSize: pagination.ParsePageSizeParam(c.DefaultQuery("pageSize", "50")),
	})
	if err != nil {
		switch err {
		case folderssvc.ErrInvalidInput:
			apierror.InvalidPayload(c)
		case folderssvc.ErrNotFound:
			apierror.Write(c, http.StatusNotFound, "folder_not_found", "Folder not found", nil)
		default:
			_ = c.Error(errs.WithStack(err))
			apierror.Internal(c, "List entries failed")
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func APIMoveEntries(folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		var req moveEntriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		if err := folderService.MoveEntries(c.Request.Context(), folderssvc.MoveEntriesInput{
			UserID:         userID.(string),
			TargetFolderID: req.TargetFolderID,
			FileIDs:        req.FileIDs,
			FolderIDs:      req.FolderIDs,
		}); err != nil {
			switch err {
			case folderssvc.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case folderssvc.ErrInvalidMove:
				apierror.Write(c, http.StatusBadRequest, "invalid_move", "Invalid move", nil)
			case folderssvc.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "entry_not_found", "Entry not found", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Move failed")
			}
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func APIDeleteEntries(folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		var req deleteEntriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		result, err := folderService.DeleteEntries(c.Request.Context(), folderssvc.DeleteEntriesInput{
			UserID:    userID.(string),
			FileIDs:   req.FileIDs,
			FolderIDs: req.FolderIDs,
		})
		if err != nil {
			switch {
			case errors.Is(err, folderssvc.ErrInvalidInput):
				apierror.InvalidPayload(c)
			case errors.Is(err, folderssvc.ErrNotFound):
				apierror.Write(c, http.StatusNotFound, "entry_not_found", "Entry not found", nil)
			case errors.Is(err, filessvc.ErrUploadCancelled):
				apierror.Write(c, http.StatusConflict, "upload_in_progress", "Upload in progress", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Delete failed")
			}
			return
		}

		c.JSON(http.StatusOK, deleteEntriesResponse{
			DeletedFiles:   result.DeletedFiles,
			DeletedFolders: result.DeletedFolders,
		})
	}
}

func APIRenameEntry(folderService *folderssvc.Service, filesService *filessvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		var req renameEntryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		encryptedMetadata, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedMetadata))
		if err != nil || len(encryptedMetadata) == 0 {
			apierror.InvalidPayload(c)
			return
		}

		switch strings.TrimSpace(req.Type) {
		case "file":
			searchTokens, err := decodeSearchTokens(req.SearchTokens)
			if err != nil {
				apierror.InvalidPayload(c)
				return
			}
			if err := filesService.RenameFile(c.Request.Context(), filessvc.RenameFileInput{
				UserID:            userID.(string),
				FileID:            strings.TrimSpace(req.ID),
				EncryptedMetadata: encryptedMetadata,
				SearchTokens:      searchTokens,
			}); err != nil {
				switch {
				case errors.Is(err, filessvc.ErrInvalidInput):
					apierror.InvalidPayload(c)
				case errors.Is(err, filessvc.ErrNotFound):
					apierror.Write(c, http.StatusNotFound, "entry_not_found", "Entry not found", nil)
				case errors.Is(err, filessvc.ErrUploadCancelled):
					apierror.Write(c, http.StatusConflict, "upload_in_progress", "Upload in progress", nil)
				default:
					_ = c.Error(errs.WithStack(err))
					apierror.Internal(c, "Rename failed")
				}
				return
			}
		case "folder":
			encryptedName, decodeErr := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedName))
			if decodeErr != nil || len(encryptedName) == 0 {
				apierror.InvalidPayload(c)
				return
			}
			if err := folderService.RenameFolder(c.Request.Context(), folderssvc.RenameFolderInput{
				UserID:            userID.(string),
				FolderID:          strings.TrimSpace(req.ID),
				EncryptedName:     encryptedName,
				EncryptedMetadata: encryptedMetadata,
			}); err != nil {
				switch {
				case errors.Is(err, folderssvc.ErrInvalidInput):
					apierror.InvalidPayload(c)
				case errors.Is(err, folderssvc.ErrNotFound):
					apierror.Write(c, http.StatusNotFound, "entry_not_found", "Entry not found", nil)
				default:
					_ = c.Error(errs.WithStack(err))
					apierror.Internal(c, "Rename failed")
				}
				return
			}
		default:
			apierror.InvalidPayload(c)
			return
		}

		c.Status(http.StatusNoContent)
	}
}
