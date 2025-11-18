package auth

import (
	"context"

	commonerrors "github.com/vaintrub/go-ddd-template/internal/common/errors"
)

// Firebase Auth middleware removed - using HttpMockMiddleware instead
// TODO: Implement production-ready auth when needed (OAuth2, custom JWT, etc.)

type User struct {
	UUID  string
	Email string
	Role  string

	DisplayName string
}

type ctxKey int

const (
	userContextKey ctxKey = iota
)

var (
	// if we expect that the user of the function may be interested with concrete error,
	// it's a good idea to provide variable with this error
	NoUserInContextError = commonerrors.NewAuthorizationError("no user in context", "no-user-found")
)

func UserFromCtx(ctx context.Context) (User, error) {
	u, ok := ctx.Value(userContextKey).(User)
	if ok {
		return u, nil
	}

	return User{}, NoUserInContextError
}
