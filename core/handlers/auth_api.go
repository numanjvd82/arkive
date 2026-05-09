package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/pkg/cookies"
	"arkive/pkg/errs"
)

type apiLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type apiUnlockRequest struct {
	Password string `json:"password"`
}

func APILogin(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req apiLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		result, validationErrors, err := svc.LoginAndLoadVault(c.Request.Context(), req.Email, req.Password, c.ClientIP())
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
			return
		}
		if err != nil {
			switch err {
			case auth.ErrVaultNotConfigured:
				c.JSON(http.StatusConflict, gin.H{"error": "vault is not configured"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
			}
			return
		}

		cookies.SetSession(c, result.SessionID, result.ExpiresAt, false)
		c.JSON(http.StatusOK, gin.H{
			"salt":               base64.StdEncoding.EncodeToString(result.VaultSalt),
			"encryptedMasterKey": base64.StdEncoding.EncodeToString(result.EncryptedMasterKey),
			"redirectTo":         "/dashboard",
		})
	}
}

func APIUnlockVault(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req apiUnlockRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		result, validationErrors, err := svc.UnlockVaultWithSession(c.Request.Context(), userIDStr, req.Password)
		if validationErrors != nil && validationErrors.HasAny() {
			c.JSON(http.StatusBadRequest, gin.H{"errors": validationErrors})
			return
		}
		if err != nil {
			switch err {
			case auth.ErrUnauthorized:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			case auth.ErrVaultNotConfigured:
				c.JSON(http.StatusConflict, gin.H{"error": "vault is not configured"})
			default:
				_ = c.Error(errs.WithStack(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "unlock failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"salt":               base64.StdEncoding.EncodeToString(result.VaultSalt),
			"encryptedMasterKey": base64.StdEncoding.EncodeToString(result.EncryptedMasterKey),
		})
	}
}
