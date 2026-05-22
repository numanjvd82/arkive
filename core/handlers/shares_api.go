package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	"arkive/core/services/shares"
	"arkive/pkg/apierror"
	"arkive/pkg/errs"
	"arkive/pkg/validation"
)

type shareCreateRequest struct {
	Token                    string `json:"token"`
	ExpiresAt                string `json:"expiresAt"`
	Password                 string `json:"password"`
	EncryptedShareKey        string `json:"encryptedShareKey"`
	EncryptedFileKeyForShare string `json:"encryptedFileKeyForShare"`
}

type shareUpdateRequest struct {
	ExpiresAt       string `json:"expiresAt"`
	Password        string `json:"password"`
	RequirePassword bool   `json:"requirePassword"`
}

func APICreateShare(svc *shares.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := strings.TrimSpace(c.Param("id"))
		if fileID == "" {
			apierror.InvalidPayload(c)
			return
		}

		var req shareCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		var expiresAt *time.Time
		if strings.TrimSpace(req.ExpiresAt) != "" {
			parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
			if err != nil {
				errors := validation.New()
				errors.Add("expiresAt", "expiry must be RFC3339")
				apierror.Validation(c, errors)
				return
			}
			expiresAt = &parsed
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}
		if strings.TrimSpace(req.EncryptedShareKey) != "" {
			if _, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedShareKey)); err != nil {
				apierror.InvalidPayload(c)
				return
			}
		}
		if strings.TrimSpace(req.EncryptedFileKeyForShare) != "" {
			if _, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedFileKeyForShare)); err != nil {
				apierror.InvalidPayload(c)
				return
			}
		}

		share, validationErrors, err := svc.CreateShare(c.Request.Context(), shares.CreateInput{
			FileID:                   fileID,
			OwnerUserID:              userID.(string),
			Token:                    req.Token,
			Password:                 req.Password,
			ExpiresAt:                expiresAt,
			EncryptedShareKey:        req.EncryptedShareKey,
			EncryptedFileKeyForShare: req.EncryptedFileKeyForShare,
		})
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				apierror.Unauthorized(c)
			case shares.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "file_not_found", "File not found", nil)
			case shares.ErrShareExists:
				apierror.Write(c, http.StatusConflict, "share_already_exists", "Share already exists", nil)
			case shares.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case shares.ErrPasswordHashFailed:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Password hashing failed")
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Share create failed")
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			apierror.Validation(c, validationErrors)
			return
		}

		status, expired := shareStatus(share)
		c.JSON(http.StatusCreated, gin.H{
			"id":                share.ID,
			"fileId":            share.FileID,
			"token":             share.Token,
			"encryptedShareKey": base64.StdEncoding.EncodeToString(share.EncryptedShareKey),
			"expiresAt":         share.ExpiresAt,
			"hasPassword":       share.PasswordHash != nil,
			"status":            status,
			"isExpired":         expired,
			"createdAt":         share.CreatedAt,
		})
	}
}

func APIUpdateShare(svc *shares.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		shareID := strings.TrimSpace(c.Param("id"))
		if shareID == "" {
			apierror.InvalidPayload(c)
			return
		}

		var req shareUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			apierror.InvalidPayload(c)
			return
		}

		var expiresAt *time.Time
		if strings.TrimSpace(req.ExpiresAt) != "" {
			parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
			if err != nil {
				errors := validation.New()
				errors.Add("expiresAt", "expiry must be RFC3339")
				apierror.Validation(c, errors)
				return
			}
			expiresAt = &parsed
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		share, validationErrors, err := svc.UpdateShareForUser(c.Request.Context(), shareID, userID.(string), expiresAt, req.Password, req.RequirePassword)
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				apierror.Unauthorized(c)
			case shares.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			case shares.ErrInvalidInput:
				apierror.InvalidPayload(c)
			case shares.ErrPasswordHashFailed:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Password hashing failed")
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Share update failed")
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			apierror.Validation(c, validationErrors)
			return
		}

		status, expired := shareStatus(share)
		c.JSON(http.StatusOK, gin.H{
			"id":                share.ID,
			"fileId":            share.FileID,
			"token":             share.Token,
			"encryptedShareKey": base64.StdEncoding.EncodeToString(share.EncryptedShareKey),
			"expiresAt":         share.ExpiresAt,
			"hasPassword":       share.PasswordHash != nil,
			"status":            status,
			"isExpired":         expired,
			"createdAt":         share.CreatedAt,
		})
	}
}

func APIGetShareForFile(svc *shares.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := strings.TrimSpace(c.Param("id"))
		if fileID == "" {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		share, err := svc.GetShareForFileForUser(c.Request.Context(), fileID, userID.(string))
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				apierror.Unauthorized(c)
			case shares.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			}
			return
		}

		status, expired := shareStatus(share)
		c.JSON(http.StatusOK, gin.H{
			"id":                share.ID,
			"fileId":            share.FileID,
			"token":             share.Token,
			"encryptedShareKey": base64.StdEncoding.EncodeToString(share.EncryptedShareKey),
			"expiresAt":         share.ExpiresAt,
			"hasPassword":       share.PasswordHash != nil,
			"status":            status,
			"isExpired":         expired,
			"createdAt":         share.CreatedAt,
		})
	}
}

func APIGetShareCryptoRecord(svc *shares.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		shareID := strings.TrimSpace(c.Param("id"))
		if shareID == "" {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		share, err := svc.GetShareForUser(c.Request.Context(), shareID, userID.(string))
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				apierror.Unauthorized(c)
			case shares.ErrNotFound:
				apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			case shares.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Share lookup failed")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":                share.ID,
			"token":             share.Token,
			"encryptedShareKey": base64.StdEncoding.EncodeToString(share.EncryptedShareKey),
			"hasPassword":       share.PasswordHash != nil,
		})
	}
}

func APIDeleteShare(svc *shares.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		shareID := strings.TrimSpace(c.Param("id"))
		if shareID == "" {
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}

		deleted, err := svc.DeleteShareForUser(c.Request.Context(), shareID, userID.(string))
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				apierror.Unauthorized(c)
			case shares.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Share delete failed")
			}
			return
		}
		if !deleted {
			apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			return
		}

		c.Status(http.StatusNoContent)
	}
}

func shareStatus(share models.Share) (string, bool) {
	if share.Status == shares.ShareStatusRevoked {
		return shares.ShareStatusRevoked, false
	}
	if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
		return "expired", true
	}
	return shares.ShareStatusActive, false
}
