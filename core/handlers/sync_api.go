package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	syncsvc "arkive/core/services/sync"
	appcontext "arkive/pkg/context"

	"arkive/pkg/apierror"
	"arkive/pkg/errs"
)

func APIListSyncEntries(syncService *syncsvc.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
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
		limit, err := parseSyncLimit(c.Query("limit"))
		if err != nil {
			apierror.InvalidPayload(c)
			return
		}
		cursor, err := parseSyncCursor(c.Query("cursor"))
		if err != nil {
			apierror.InvalidPayload(c)
			return
		}
		if !hasFolderID {
			folderID = nil
		}

		result, err := syncService.ListEntries(c.Request.Context(), models.ListEntriesPageInput{
			UserID:         user.ID,
			FolderID:       folderID,
			IncludeDeleted: includeDeleted,
			Limit:          limit,
			Cursor:         cursor,
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

		c.JSON(http.StatusOK, models.SyncEntriesResponse{
			Entries:    result.Entries,
			NextCursor: result.NextCursor,
			HasMore:    result.HasMore,
		})
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

func parseSyncLimit(raw string) (int, error) {
	const (
		defaultLimit = 100
		maxLimit     = 500
	)

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	if limit <= 0 {
		return 0, strconv.ErrSyntax
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit, nil
}

func parseSyncCursor(raw string) (*models.SyncEntriesCursor, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	decoded, err := base64.RawURLEncoding.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}

	var payload struct {
		UpdatedAt string `json:"updated_at"`
		Type      string `json:"type"`
		ID        string `json:"id"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, err
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, payload.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &models.SyncEntriesCursor{
		UpdatedAt: updatedAt,
		Type:      payload.Type,
		ID:        payload.ID,
	}, nil
}
