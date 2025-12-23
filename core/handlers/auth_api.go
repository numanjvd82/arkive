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
		user, err := svc.Authenticate(c.Request.Context(), req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		accessToken, accessExpires, err := svc.CreateAccessToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token creation failed"})
			return
		}
		refreshToken, refreshExpires, err := svc.CreateRefreshToken(c.Request.Context(), user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token creation failed"})
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
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		newRefreshToken, refreshExpires, userID, err := svc.RotateRefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		accessToken, accessExpires, err := svc.CreateAccessToken(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token creation failed"})
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
		if req.RefreshToken != "" {
			_ = svc.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
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
