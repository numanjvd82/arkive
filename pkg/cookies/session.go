package cookies

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const SessionName = "arkive_session"

func SetSession(c *gin.Context, sessionID string, expiresAt time.Time, secure bool) {
	SetCustom(c, SessionName, sessionID, expiresAt, secure)
}

func ClearSession(c *gin.Context, secure bool) {
	ClearCustom(c, SessionName, secure)
}

func SetCustom(c *gin.Context, name, value string, expiresAt time.Time, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func ClearCustom(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
