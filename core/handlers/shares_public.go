package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"arkive/core/models"
	filessvc "arkive/core/services/files"
	"arkive/core/services/shares"
	"arkive/core/web"
	"arkive/core/web/pages"
	"arkive/pkg/apierror"
	"arkive/pkg/cookies"
	"arkive/pkg/errs"
)

const (
	shareAccessCookieName = "arkive_share_access"
	shareAccessTTL        = 15 * time.Minute
)

func PublicShareView(shareService *shares.Service, filesService *filessvc.Service, cookieSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Param("token"))
		if token == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		share, err := shareService.GetShareByToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if share.Status != shares.ShareStatusActive {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := filesService.GetFileForShare(c.Request.Context(), share.FileID)
		if err != nil {
			switch err {
			case filessvc.ErrNotFound, filessvc.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case filessvc.ErrInvalidInput:
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				_ = c.Error(errs.WithStack(err))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		if share.PasswordHash != nil && !hasShareAccess(c, share, cookieSecret) {
			web.Render(c, pages.PublicSharePassword(pages.PublicSharePasswordProps{
				Token: token,
				File:  file,
			}))
			return
		}

		renderShareLanding(c, filesService, token, share, file)
	}
}

func PublicShareUnlock(shareService *shares.Service, filesService *filessvc.Service, cookieSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Param("token"))
		if token == "" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		share, err := shareService.GetShareByToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if share.Status != shares.ShareStatusActive {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		file, err := filesService.GetFileForShare(c.Request.Context(), share.FileID)
		if err != nil {
			switch err {
			case filessvc.ErrNotFound, filessvc.ErrUploadCancelled:
				c.AbortWithStatus(http.StatusNotFound)
			case filessvc.ErrInvalidInput:
				c.AbortWithStatus(http.StatusBadRequest)
			default:
				_ = c.Error(errs.WithStack(err))
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}

		if share.PasswordHash == nil || hasShareAccess(c, share, cookieSecret) {
			renderShareLanding(c, filesService, token, share, file)
			return
		}

		password := strings.TrimSpace(c.PostForm("password"))
		if password == "" || bcrypt.CompareHashAndPassword([]byte(*share.PasswordHash), []byte(password)) != nil {
			if isJSONShareRequest(c) {
				apierror.Write(c, http.StatusUnauthorized, "invalid_share_password", "Password is incorrect.", nil)
				return
			}
			c.Status(http.StatusUnauthorized)
			web.Render(c, pages.PublicSharePassword(pages.PublicSharePasswordProps{
				Token:   token,
				File:    file,
				Message: "Password is incorrect.",
			}))
			return
		}

		setShareAccess(c, share, cookieSecret)
		if isJSONShareRequest(c) {
			c.Status(http.StatusNoContent)
			return
		}
		renderShareLanding(c, filesService, token, share, file)
	}
}

func APIPublicShareRecord(shareService *shares.Service, filesService *filessvc.Service, cookieSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.Param("token"))
		if token == "" {
			apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			return
		}

		share, err := shareService.GetShareByToken(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
				return
			}
			_ = c.Error(errs.WithStack(err))
			apierror.Internal(c, "Share lookup failed")
			return
		}
		if share.Status != shares.ShareStatusActive {
			apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			return
		}
		if share.PasswordHash != nil && !hasShareAccess(c, share, cookieSecret) {
			apierror.Write(c, http.StatusForbidden, "forbidden", "Share access denied", nil)
			return
		}
		if !share.AllowPreview && !share.AllowDownload {
			apierror.Write(c, http.StatusForbidden, "forbidden", "Share access denied", nil)
			return
		}
		if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
			apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			return
		}

		file, err := filesService.GetFileForShare(c.Request.Context(), share.FileID)
		if err != nil {
			switch err {
			case filessvc.ErrNotFound, filessvc.ErrUploadCancelled:
				apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
			case filessvc.ErrInvalidInput:
				apierror.InvalidPayload(c)
			default:
				_ = c.Error(errs.WithStack(err))
				apierror.Internal(c, "Share lookup failed")
			}
			return
		}

		sourceURL, err := filesService.PresignShareSourceForFile(c.Request.Context(), file)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			apierror.Internal(c, "Share source failed")
			return
		}

		record, err := shareService.GetPublicShareRecord(c.Request.Context(), token)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				apierror.Write(c, http.StatusNotFound, "share_not_found", "Share not found", nil)
				return
			}
			_ = c.Error(errs.WithStack(err))
			apierror.Internal(c, "Share record failed")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"shareId":                  record.ShareID,
			"token":                    record.Token,
			"fileId":                   record.FileID,
			"vaultId":                  record.VaultID,
			"encryptionVersion":        record.EncryptionVersion,
			"chunkSize":                record.ChunkSize,
			"totalChunks":              record.TotalChunks,
			"plaintextSize":            record.PlaintextSize,
			"encryptedHash":            base64.StdEncoding.EncodeToString(record.EncryptedHash),
			"encryptedMetadata":        base64.StdEncoding.EncodeToString(record.EncryptedMetadata),
			"encryptedManifest":        base64.StdEncoding.EncodeToString(record.EncryptedManifest),
			"encryptedFileKeyForShare": base64.StdEncoding.EncodeToString(record.EncryptedFileKeyForShare),
			"sourceUrl":                sourceURL,
			"allowPreview":             share.AllowPreview,
			"allowDownload":            share.AllowDownload,
			"shareFileKeyAad":          "arkive:share-file-key:v1:" + record.FileID + ":" + record.Token,
			"metadataAad":              "arkive:file-metadata:v1:" + record.VaultID + ":" + record.FileID,
			"manifestAad":              "arkive:file-manifest:v1:" + record.VaultID + ":" + record.FileID,
		})
	}
}

func renderShareLanding(c *gin.Context, filesService *filessvc.Service, token string, share models.Share, file models.File) {
	viewURL := ""
	isImage := false
	isVideo := false
	viewable := false
	downloadURL, err := filesService.PresignShareDownloadForFile(c.Request.Context(), file)
	if err != nil {
		_ = c.Error(errs.WithStack(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if !share.AllowDownload {
		downloadURL = "#"
	}

	shareURL := buildShareURL(c, token)
	web.Render(c, pages.PublicShareViewPage(pages.PublicShareViewProps{
		Token:       token,
		File:        file,
		ViewURL:     viewURL,
		DownloadURL: downloadURL,
		IsImage:     isImage,
		IsVideo:     isVideo,
		Viewable:    viewable && viewURL != "",
		ShareURL:    shareURL,
		SharedAt:    share.CreatedAt,
	}))
}

func hasShareAccess(c *gin.Context, share models.Share, cookieSecret string) bool {
	if share.PasswordHash == nil {
		return true
	}
	cookie, err := c.Request.Cookie(shareAccessCookieName)
	if err != nil || cookie == nil {
		return false
	}
	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}
	payload := string(payloadBytes)
	payloadParts := strings.Split(payload, ":")
	if len(payloadParts) != 2 {
		return false
	}
	if payloadParts[0] != share.Token {
		return false
	}
	expiresUnix, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil || expiresUnix <= time.Now().Unix() {
		return false
	}
	expected := signShareAccessPayload(payload, cookieSecret)
	return hmac.Equal([]byte(parts[1]), []byte(expected))
}

func setShareAccess(c *gin.Context, share models.Share, cookieSecret string) {
	if share.PasswordHash == nil {
		return
	}
	expiresAt := time.Now().Add(shareAccessTTL)
	payload := share.Token + ":" + strconv.FormatInt(expiresAt.Unix(), 10)
	value := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signShareAccessPayload(payload, cookieSecret)
	cookies.SetCustom(c, shareAccessCookieName, value, expiresAt, isSecureRequest(c))
}

func signShareAccessPayload(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func isJSONShareRequest(c *gin.Context) bool {
	return strings.EqualFold(strings.TrimSpace(c.GetHeader("X-Requested-With")), "XMLHttpRequest")
}

func isSecureRequest(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	forwarded := strings.ToLower(strings.TrimSpace(strings.Split(c.GetHeader("X-Forwarded-Proto"), ",")[0]))
	return forwarded == "https"
}

func buildShareURL(c *gin.Context, token string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.ToLower(strings.TrimSpace(strings.Split(forwarded, ",")[0]))
	}
	host := c.Request.Host
	return scheme + "://" + host + "/s/" + token
}
