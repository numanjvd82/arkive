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
