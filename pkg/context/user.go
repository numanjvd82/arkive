package context

import (
	"github.com/gin-gonic/gin"

	"arkive/core/models"
	"arkive/core/services/auth"
	"arkive/pkg/cookies"
)

const UserKey = "user"

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
		cookies.ClearSession(c, false)
		return models.User{}, false, err
	}

	c.Set(UserKey, user)
	return user, true, nil
}

func UserFromContext(c *gin.Context) (models.User, bool) {
	userValue, ok := c.Get(UserKey)
	if !ok {
		return models.User{}, false
	}
	user, ok := userValue.(models.User)
	return user, ok
}
