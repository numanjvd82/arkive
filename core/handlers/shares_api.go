package handlers

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/models"
	"arkive/core/services/shares"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var req shareCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var expiresAt *time.Time
		if strings.TrimSpace(req.ExpiresAt) != "" {
			parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
			if err != nil {
				errors := validation.New()
				errors.Add("expiresAt", "expiry must be RFC3339")
				c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
				return
			}
			expiresAt = &parsed
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if strings.TrimSpace(req.EncryptedShareKey) != "" {
			if _, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedShareKey)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
				return
			}
		}
		if strings.TrimSpace(req.EncryptedFileKeyForShare) != "" {
			if _, err := base64.StdEncoding.DecodeString(strings.TrimSpace(req.EncryptedFileKeyForShare)); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
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
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case shares.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			case shares.ErrShareExists:
				c.JSON(http.StatusConflict, gin.H{"error": "share already exists"})
			case shares.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case shares.ErrPasswordHashFailed:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "share create failed"})
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var req shareUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var expiresAt *time.Time
		if strings.TrimSpace(req.ExpiresAt) != "" {
			parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpiresAt))
			if err != nil {
				errors := validation.New()
				errors.Add("expiresAt", "expiry must be RFC3339")
				c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
				return
			}
			expiresAt = &parsed
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		share, validationErrors, err := svc.UpdateShareForUser(c.Request.Context(), shareID, userID.(string), expiresAt, req.Password, req.RequirePassword)
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case shares.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			case shares.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case shares.ErrPasswordHashFailed:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "share update failed"})
			}
			return
		}
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		share, err := svc.GetShareForFileForUser(c.Request.Context(), fileID, userID.(string))
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case shares.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		share, err := svc.GetShareForUser(c.Request.Context(), shareID, userID.(string))
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case shares.ErrNotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
			case shares.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "share lookup failed"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		deleted, err := svc.DeleteShareForUser(c.Request.Context(), shareID, userID.(string))
		if err != nil {
			switch err {
			case shares.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case shares.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "share delete failed"})
			}
			return
		}
		if !deleted {
			c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
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
