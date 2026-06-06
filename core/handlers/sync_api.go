package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	syncsvc "arkive/core/services/sync"
	"arkive/pkg/apierror"
	"arkive/pkg/errs"
)

func APIListSyncEntries(syncService *syncsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		folderID, hasFolderID, err := parseSyncFolderID(c.Query("folder_id"))
		if err != nil {
			apierror.InvalidPayload(c)
			return
		}
		includeDeleted, err := parseSyncIncludeDeleted(c.Query("include_deleted"))
		if err != nil {
			apierror.InvalidPayload(c)
			return
		}
		if !hasFolderID {
			folderID = nil
		}

		result, err := syncService.ListEntries(c.Request.Context(), syncsvc.ListEntriesInput{
			UserID:         userID.(string),
			FolderID:       folderID,
			IncludeDeleted: includeDeleted,
		})
		if err != nil {
			switch err {
			case syncsvc.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case syncsvc.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "folder_not_found", "Folder not found", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Sync listing failed")
			}
			return
		}

		c.JSON(http.StatusOK, models.SyncEntriesResponse{Entries: result.Entries})
	}
}

func parseSyncFolderID(raw string) (*string, bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, "null") {
		return nil, trimmed != "", nil
	}
	return &trimmed, true, nil
}

func parseSyncIncludeDeleted(raw string) (bool, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false, nil
	}
	return strconv.ParseBool(trimmed)
}
