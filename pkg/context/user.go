package context

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"arkive/core/models"
	"arkive/core/services/auth"
	"arkive/pkg/cookies"
)

const userKey = "user"

func LoadUser(c *gin.Context, svc *auth.Service) (models.User, bool, error) {
	cookie, err := c.Request.Cookie(cookies.SessionName)
	if err != nil || cookie.Value == "" {
		return models.User{}, false, nil
	}

	userID, err := svc.ValidateSession(c.Request.Context(), cookie.Value)
	if err != nil {
		cookies.ClearSession(c, false)
		return models.User{}, false, nil
	}

	user, err := svc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			cookies.ClearSession(c, false)
			return models.User{}, false, nil
		}
		return models.User{}, false, err
	}

	c.Set(userKey, user)
	return user, true, nil
}
