package cookies

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const SessionName = "arkive_session"

func SetSession(c *gin.Context, sessionID string, expiresAt time.Time, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func ClearSession(c *gin.Context, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     SessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}
