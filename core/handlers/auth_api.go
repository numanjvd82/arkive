package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func APILogin(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		accessToken, accessExpires, refreshToken, refreshExpires, err := svc.LoginTokens(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			switch err {
			case auth.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case auth.ErrInvalidCredentials:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "token creation failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"access_expires": accessExpires,
			"refresh_token": refreshToken,
			"refresh_expires": refreshExpires,
		})
	}
}

func APIRefresh(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req refreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		accessToken, accessExpires, newRefreshToken, refreshExpires, err := svc.RefreshTokens(c.Request.Context(), req.RefreshToken)
		if err != nil {
			switch err {
			case auth.ErrInvalidInput:
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			case auth.ErrRefreshTokenInvalid:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "token creation failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"access_expires": accessExpires,
			"refresh_token": newRefreshToken,
			"refresh_expires": refreshExpires,
		})
	}
}

func APILogout(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req refreshRequest
		_ = c.ShouldBindJSON(&req)
		if err := svc.RevokeRefreshToken(c.Request.Context(), req.RefreshToken); err != nil {
			if err != auth.ErrInvalidInput && err != auth.ErrRefreshTokenInvalid {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
				return
			}
		}
		c.Status(http.StatusNoContent)
	}
}

func APIMe(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, err := svc.GetUserByID(c.Request.Context(), userID.(string))
		if err != nil {
			if err == auth.ErrInvalidInput {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         user.ID,
			"brand_name": user.BrandName,
			"email":      user.Email,
		})
	}
}
