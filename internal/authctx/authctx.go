package authctx

import "context"

type contextKey string

const userIDKeyContext contextKey = "userID"
const roleKeyContext contextKey = "role"

func WithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userIDKeyContext, userID)
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	if id, ok := ctx.Value(userIDKeyContext).(int); ok {
		return id, false
	}
	return 0, true
}

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, roleKeyContext, role)
}

func RoleFromContext(ctx context.Context) (string, bool) {
	if role, ok := ctx.Value(roleKeyContext).(string); ok {
		return role, false
	}
	return "", true
}
