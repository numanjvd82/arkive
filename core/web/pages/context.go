package pages

import "arkive/core/models"

type PageContext struct {
	User *models.User
}

func ContextWithUser(user models.User) PageContext {
	return PageContext{User: &user}
}
