package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	filessvc "arkive/core/services/files"
	folderssvc "arkive/core/services/folders"
	"arkive/pkg/apierror"
	appcontext "arkive/pkg/context"
)

type searchResult struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Title    string `json:"title"`
	Meta     string `json:"meta,omitempty"`
	Status   string `json:"status,omitempty"`
	URL      string `json:"url"`
	Category string `json:"category"`
}

type encryptedSearchFileResult struct {
	ID                string `json:"id"`
	Kind              string `json:"kind"`
	VaultID           string `json:"vaultId"`
	EncryptedMetadata string `json:"encryptedMetadata"`
	EncryptedFileKey  string `json:"encryptedFileKey"`
	Score             int64  `json:"score"`
	URL               string `json:"url"`
}

type encryptedSearchFolderResult struct {
	ID                string  `json:"id"`
	Kind              string  `json:"kind"`
	ParentFolderID    *string `json:"parentFolderId"`
	VaultID           string  `json:"vaultId"`
	EncryptedName     string  `json:"encryptedName"`
	EncryptedMetadata string  `json:"encryptedMetadata"`
	Score             int64   `json:"score"`
	URL               string  `json:"url"`
}

type searchRequest struct {
	Tokens []string `json:"tokens"`
	Limit  int      `json:"limit"`
}

func APISearch(filesService *filessvc.Service, folderService *folderssvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			apierror.Unauthorized(c)
			return
		}

		var req searchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}
		if len(req.Tokens) == 0 {
			c.JSON(http.StatusOK, gin.H{"results": gin.H{}})
			return
		}

		tokenHashes := make([][]byte, 0, len(req.Tokens))
		for _, token := range req.Tokens {
			decoded, err := filessvc.DecodeSearchTokenString(token)
			if err != nil {
				apierror.InvalidPayload(c)
				return
			}
			tokenHashes = append(tokenHashes, decoded)
		}

		limit := req.Limit
		if limit <= 0 || limit > 20 {
			limit = 20
		}

		files, err := filesService.SearchCompletedUploadsByTokens(c.Request.Context(), user.ID, user.ID, tokenHashes, limit)
		if err != nil {
			if err == filessvc.ErrInvalidInput {
				apierror.InvalidPayload(c)
				return
			}
			apierror.Internal(c, "Search failed")
			return
		}
		folders, err := folderService.SearchFoldersByTokens(c.Request.Context(), user.ID, user.ID, tokenHashes, limit)
		if err != nil {
			if err == folderssvc.ErrInvalidInput {
				apierror.InvalidPayload(c)
				return
			}
			apierror.Internal(c, "Search failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"results": gin.H{
				"folders":  mapEncryptedFolderResults(folders),
				"files":    mapEncryptedFileResults(files),
				"shares":   []searchResult{},
				"settings": []searchResult{},
			},
		})
	}
}

func mapEncryptedFolderResults(folders []models.Folder) []encryptedSearchFolderResult {
	results := make([]encryptedSearchFolderResult, 0, len(folders))
	for _, folder := range folders {
		results = append(results, encryptedSearchFolderResult{
			ID:                folder.ID,
			Kind:              "folder",
			ParentFolderID:    folder.ParentFolderID,
			VaultID:           folder.VaultID,
			EncryptedName:     base64.StdEncoding.EncodeToString(folder.EncryptedName),
			EncryptedMetadata: base64.StdEncoding.EncodeToString(folder.EncryptedMetadata),
			Score:             folder.SearchScore,
			URL:               "/folders/" + folder.ID,
		})
	}
	return results
}

func mapEncryptedFileResults(files []models.File) []encryptedSearchFileResult {
	results := make([]encryptedSearchFileResult, 0, len(files))
	for _, file := range files {
		url := "/api/files/" + file.ID + "/download"
		if isSearchPreviewable(file.UploadStatus) {
			url = "/files/" + file.ID + "/view"
		}
		results = append(results, encryptedSearchFileResult{
			ID:                file.ID,
			Kind:              "file",
			VaultID:           file.UserID,
			EncryptedMetadata: base64.StdEncoding.EncodeToString(file.EncryptedMetadata),
			EncryptedFileKey:  base64.StdEncoding.EncodeToString(file.EncryptedFileKey),
			Score:             file.SearchScore,
			URL:               url,
		})
	}
	return results
}

func isSearchPreviewable(status string) bool {
	return strings.TrimSpace(status) == "complete"
}
