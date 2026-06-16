package auth

import (
	"context"

	"github.com/zhitoo/hls_converter/internal/models"
)

func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(userKey).(*models.User)
	return u
}
