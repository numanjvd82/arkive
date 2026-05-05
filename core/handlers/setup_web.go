package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"arkive/core/services/auth"
	settingssvc "arkive/core/services/settings"
	"arkive/core/services/setup"
	"arkive/core/web"
	"arkive/core/web/pages"
	appcontext "arkive/pkg/context"
	"arkive/pkg/cookies"
	"arkive/pkg/errs"
	"arkive/pkg/validation"
)

const (
	recoveryPendingCookie = "arkive_setup_recovery_pending"
	recoveryCookieTTL     = 15 * time.Minute
)

var recoveryDemoWords = []string{
	"abandon", "justice", "degree", "canvas",
	"theory", "vivid", "script", "matrix",
	"pioneer", "quantum", "resilient", "sovereign",
	"utility", "vacuum", "weather", "yonder",
	"zenith", "alpine", "beacon", "cipher",
	"dorsal", "engine", "fossil", "glider",
}

func WebRoot(authSvc *auth.Service, setupSvc *setup.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		initialized, err := setupSvc.IsInitialized(c.Request.Context())
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !initialized {
			c.Redirect(http.StatusSeeOther, "/setup")
			return
		}

		if _, ok, err := appcontext.LoadUser(c, authSvc); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		} else if ok {
			c.Redirect(http.StatusSeeOther, "/dashboard")
			return
		}

		c.Redirect(http.StatusSeeOther, "/login")
	}
}

func WebSetupGet(svc *setup.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		initialized, err := svc.IsInitialized(c.Request.Context())
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if initialized {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		web.Render(c, pages.SetupPage(pages.SetupPageProps{
			Ctx: pages.PageContext{},
		}))
	}
}

func WebSetupPost(svc *setup.Service) gin.HandlerFunc {
	type setupForm struct {
		BrandName         string `form:"brand_name"`
		Email             string `form:"email"`
		Password          string `form:"password"`
		ConfirmPassword   string `form:"confirm_password"`
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
		var form setupForm
		if err := c.ShouldBind(&form); err != nil {
			renderSetupForm(c, form, validation.Errors{
				validation.GeneralKey: "Please fill out the form.",
			})
			return
		}

		storageSettings, validationErrors := settingssvc.BuildStorageSettings(settingssvc.StorageInput{
			Provider:          form.StorageProvider,
			LocalPath:         form.LocalPath,
			StorageGB:         form.StorageGB,
			S3AccessKeyID:     form.S3AccessKeyID,
			S3SecretAccessKey: form.S3SecretAccessKey,
			S3SessionToken:    form.S3SessionToken,
			S3Bucket:          form.S3Bucket,
			S3Endpoint:        form.S3Endpoint,
			S3Region:          form.S3Region,
			S3UsePathStyle:    form.S3UsePathStyle,
		})
		if len(validationErrors) > 0 {
			renderSetupForm(c, form, validationErrors)
			return
		}
		user, validationErrors, err := svc.CreateInitialAdmin(c.Request.Context(), setup.InitialAdminInput{
			BrandName:       form.BrandName,
			Email:           form.Email,
			Password:        form.Password,
			ConfirmPassword: form.ConfirmPassword,
			Storage:         storageSettings,
			LocalStorageGB:  form.StorageGB,
		})
		if len(validationErrors) > 0 {
			renderSetupForm(c, form, validationErrors)
			return
		}
		if err != nil {
			if errors.Is(err, setup.ErrAlreadyInitialized) {
				c.Redirect(http.StatusSeeOther, "/login")
				return
			}
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := setRecoveryPending(c, svc, user.ID); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Redirect(http.StatusSeeOther, "/setup/recovery-key")
	}
}

func WebSetupRecoveryGet(svc *setup.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		initialized, err := svc.IsInitialized(c.Request.Context())
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !initialized {
			c.Redirect(http.StatusSeeOther, "/setup")
			return
		}
		ok, err := hasRecoveryPending(c, svc)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		web.Render(c, pages.SetupRecoveryPage(pages.SetupRecoveryPageProps{
			Ctx:   pages.PageContext{},
			Words: recoveryDemoWords,
		}))
	}
}

func WebSetupRecoveryPost(svc *setup.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		initialized, err := svc.IsInitialized(c.Request.Context())
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !initialized {
			c.Redirect(http.StatusSeeOther, "/setup")
			return
		}
		ok, err := hasRecoveryPending(c, svc)
		if err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		acknowledged := strings.TrimSpace(c.PostForm("confirm_backup")) == "true"
		if !acknowledged {
			web.Render(c, pages.SetupRecoveryPage(pages.SetupRecoveryPageProps{
				Ctx:          pages.PageContext{},
				Words:        recoveryDemoWords,
				Error:        "You must confirm that the recovery key has been securely backed up.",
				Acknowledged: false,
			}))
			return
		}

		if err := clearRecoveryPending(c, svc); err != nil {
			_ = c.Error(errs.WithStack(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Redirect(http.StatusSeeOther, "/login?msg=account-created")
	}
}

func renderSetupForm(c *gin.Context, form struct {
	BrandName         string `form:"brand_name"`
	Email             string `form:"email"`
	Password          string `form:"password"`
	ConfirmPassword   string `form:"confirm_password"`
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
}, validationErrors validation.Errors) {
	web.Render(c, pages.SetupPage(pages.SetupPageProps{
		Ctx:               pages.PageContext{},
		Errors:            validationErrors,
		BrandName:         strings.TrimSpace(form.BrandName),
		Email:             strings.TrimSpace(form.Email),
		StorageProvider:   strings.TrimSpace(form.StorageProvider),
		LocalPath:         strings.TrimSpace(form.LocalPath),
		StorageGB:         strings.TrimSpace(form.StorageGB),
		S3AccessKeyID:     strings.TrimSpace(form.S3AccessKeyID),
		S3SecretAccessKey: strings.TrimSpace(form.S3SecretAccessKey),
		S3SessionToken:    strings.TrimSpace(form.S3SessionToken),
		S3Bucket:          strings.TrimSpace(form.S3Bucket),
		S3Endpoint:        strings.TrimSpace(form.S3Endpoint),
		S3Region:          strings.TrimSpace(form.S3Region),
		S3UsePathStyle:    strings.TrimSpace(form.S3UsePathStyle) == "on",
	}))
}

func setRecoveryPending(c *gin.Context, svc *setup.Service, userID string) error {
	token, err := svc.IssueRecoverySetupToken(c.Request.Context(), userID, recoveryCookieTTL)
	if err != nil {
		return err
	}
	cookies.SetCustom(c, recoveryPendingCookie, token, time.Now().Add(recoveryCookieTTL), false)
	return nil
}

func hasRecoveryPending(c *gin.Context, svc *setup.Service) (bool, error) {
	cookie, err := c.Request.Cookie(recoveryPendingCookie)
	if err != nil {
		return false, nil
	}
	token := strings.TrimSpace(cookie.Value)
	if token == "" {
		return false, nil
	}
	return svc.HasValidRecoverySetupToken(c.Request.Context(), token)
}

func clearRecoveryPending(c *gin.Context, svc *setup.Service) error {
	if cookie, err := c.Request.Cookie(recoveryPendingCookie); err == nil {
		token := strings.TrimSpace(cookie.Value)
		if token != "" {
			if err := svc.ClearRecoverySetupToken(c.Request.Context(), token); err != nil {
				return err
			}
		}
	}
	cookies.ClearCustom(c, recoveryPendingCookie, false)
	return nil
}
