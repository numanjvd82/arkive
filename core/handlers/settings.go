package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"arkive/core/database"
	"arkive/core/models"
	settingsrepo "arkive/core/repositories/settings"
	usersrepo "arkive/core/repositories/users"
	"arkive/core/services/setup"
	"arkive/core/services/uploads"
	"arkive/core/web"
	pages "arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/errs"
	"arkive/pkg/validation"
)

func WebSettings(uploadService *uploads.Service, settingsRepo *settingsrepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if err := uploadService.TouchUserActivity(c.Request.Context(), user.ID); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		storageSettings, _ := settingsRepo.GetStorageSettings(c.Request.Context(), uploadService.DB())
		web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
			Ctx:             pages.ContextWithUser(user),
			StorageSettings: storageSettings,
		}))
	}
}

func WebSettingsStoragePost(db database.PgPool, settingsRepo *settingsrepo.Repository, userRepo *usersrepo.Repository) gin.HandlerFunc {
	type storageForm struct {
		StorageProvider   string `form:"storage_provider"`
		LocalPath         string `form:"local_path"`
		StorageGB         string `form:"storage_gb"`
		S3AccessKeyID     string `form:"s3_access_key_id"`
		S3SecretAccessKey string `form:"s3_secret_access_key"`
		S3SessionToken    string `form:"s3_session_token"`
		S3Bucket          string `form:"s3_bucket"`
		S3Endpoint        string `form:"s3_endpoint"`
		S3Region          string `form:"s3_region"`
		S3UsePathStyle    string `form:"s3_use_path_style"`
	}

	return func(c *gin.Context) {
		user, ok := appcontext.UserFromContext(c)
		if !ok || user.ID == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var form storageForm
		if err := c.ShouldBind(&form); err != nil {
			renderSettingsStorage(c, user, models.StorageSettings{}, validation.Errors{validation.GeneralKey: "Please fill out the form."})
			return
		}

		current, _ := settingsRepo.GetStorageSettings(c.Request.Context(), db)
		settings, validationErrors := setup.BuildStorageSettings(
			form.StorageProvider,
			form.LocalPath,
			form.StorageGB,
			form.S3AccessKeyID,
			form.S3SecretAccessKey,
			form.S3SessionToken,
			form.S3Bucket,
			form.S3Endpoint,
			form.S3Region,
			form.S3UsePathStyle,
		)
		if settings.Provider == "s3" && strings.TrimSpace(settings.S3SecretAccessKey) == "" {
			settings.S3SecretAccessKey = current.S3SecretAccessKey
		}
		setup.ValidateStorageSettings(settings, validationErrors)
		if len(validationErrors) > 0 {
			renderSettingsStorage(c, user, settings, validationErrors)
			return
		}

		tx, err := db.BeginTx(c.Request.Context(), pgx.TxOptions{})
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		defer func() { _ = tx.Rollback(c.Request.Context()) }()

		if err := settingsRepo.SaveStorageSettings(c.Request.Context(), tx, settings); err != nil {
			if errors.Is(err, settingsrepo.ErrStorageSettingsNotFound) {
				c.Status(http.StatusNotFound)
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		quotaBytes := settings.MaxStorageBytes
		if quotaBytes == 0 {
			quotaBytes = 9223372036854775807
		}
		if err := userRepo.UpdateQuota(c.Request.Context(), tx, user.ID, quotaBytes); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(c.Request.Context()); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusSeeOther, "/settings?msg=storage-updated")
	}
}

func renderSettingsStorage(c *gin.Context, user models.User, settings models.StorageSettings, validationErrors validation.Errors) {
	web.Render(c, pages.SettingsPage(pages.SettingsPageProps{
		Ctx:             pages.ContextWithUser(user),
		StorageSettings: settings,
		StorageGB:       storageGB(settings.MaxStorageBytes),
		Errors:          validationErrors,
	}))
}

func storageGB(bytes int64) string {
	if bytes <= 0 || bytes == 9223372036854775807 {
		return "0"
	}
	return strconv.FormatInt(bytes/(1024*1024*1024), 10)
}
