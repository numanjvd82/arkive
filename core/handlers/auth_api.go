package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	"arkive/pkg/apierror"
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
			apierror.InvalidPayload(c)
			return
		}

		result, validationErrors, err := svc.LoginAndLoadVault(c.Request.Context(), req.Email, req.Password, c.ClientIP())
		if validationErrors != nil && validationErrors.HasAny() {
			apierror.Validation(c, validationErrors)
			return
		}
		if err != nil {
			switch err {
			case auth.ErrVaultNotConfigured:
				apierror.Write(c, http.StatusConflict, "vault_not_configured", "Vault is not configured", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Login failed")
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
			apierror.InvalidPayload(c)
			return
		}

		userID, ok := c.Get("user_id")
		if !ok {
			apierror.Unauthorized(c)
			return
		}
		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			apierror.Unauthorized(c)
			return
		}

		result, validationErrors, err := svc.UnlockVaultWithSession(c.Request.Context(), userIDStr, req.Password)
		if validationErrors != nil && validationErrors.HasAny() {
			apierror.Validation(c, validationErrors)
			return
		}
		if err != nil {
			switch err {
			case auth.ErrUnauthorized:
				apierror.Unauthorized(c)
			case auth.ErrVaultNotConfigured:
				apierror.Write(c, http.StatusConflict, "vault_not_configured", "Vault is not configured", nil)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Unlock failed")
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"salt":               base64.StdEncoding.EncodeToString(result.VaultSalt),
			"encryptedMasterKey": base64.StdEncoding.EncodeToString(result.EncryptedMasterKey),
		})
	}
}
