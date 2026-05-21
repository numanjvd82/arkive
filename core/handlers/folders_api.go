package handlers

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	filessvc "arkive/core/services/files"
	folderssvc "arkive/core/services/folders"
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
	Type              string `json:"type"`
	ID                string `json:"id"`
	EncryptedName     string `json:"encryptedName"`
	EncryptedMetadata string `json:"encryptedMetadata"`
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req createFolderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		encryptedName, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedName))
		if err != nil || len(encryptedName) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var encryptedMetadata []byte
		if strings.TrimSpace(req.EncryptedMetadata) != "" {
			encryptedMetadata, err = base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedMetadata))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
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
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case folderssvc.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "parent folder not found"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "create folder failed"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
			return
		}
		apiListFolderEntries(c, folderService, &folderID)
	}
}

func apiListFolderEntries(c *gin.Context, folderService *folderssvc.Service, folderID *string) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		case folderssvc.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			_ = c.Error(errs.WithStack(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "list entries failed"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func APIMoveEntries(folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req moveEntriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
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
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case folderssvc.ErrInvalidMove:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid move"})
			case folderssvc.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "move failed"})
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req deleteEntriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
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
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case errors.Is(err, folderssvc.ErrNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			case errors.Is(err, filessvc.ErrUploadCancelled):
				c.JSON(http.StatusConflict, gin.H{"error": "upload in progress"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var req renameEntryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		encryptedMetadata, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedMetadata))
		if err != nil || len(encryptedMetadata) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		switch strings.TrimSpace(req.Type) {
		case "file":
			if err := filesService.RenameFile(c.Request.Context(), filessvc.RenameFileInput{
				UserID:            userID.(string),
				FileID:            strings.TrimSpace(req.ID),
				EncryptedMetadata: encryptedMetadata,
			}); err != nil {
				switch {
				case errors.Is(err, filessvc.ErrInvalidInput):
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
				case errors.Is(err, filessvc.ErrNotFound):
					c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
				case errors.Is(err, filessvc.ErrUploadCancelled):
					c.JSON(http.StatusConflict, gin.H{"error": "upload in progress"})
				default:
					_ = c.Error(errs.WithStack(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "rename failed"})
				}
				return
			}
		case "folder":
			encryptedName, decodeErr := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedName))
			if decodeErr != nil || len(encryptedName) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
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
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
				case errors.Is(err, folderssvc.ErrNotFound):
					c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
				default:
					_ = c.Error(errs.WithStack(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "rename failed"})
				}
				return
			}
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}
