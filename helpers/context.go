package helpers

import (
	"context"

	"async/models"
)

type userCtxKey struct{}

func ContextWithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, user)
}

func UserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userCtxKey{}).(*models.User)
	if !ok || user == nil {
		return nil, false
	}
	return user, true
}
