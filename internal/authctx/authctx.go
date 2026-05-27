package authctx

import (
	"context"

	"github.com/Alex-Blacks/Purchases/internal/policy"
)

type contextKey string

const actorKeyContext contextKey = "actor"

func WithActor(ctx context.Context, actor policy.Actor) context.Context {
	return context.WithValue(ctx, actorKeyContext, actor)
}

func ActorFromContext(ctx context.Context) (policy.Actor, bool) {
	if actor, ok := ctx.Value(actorKeyContext).(policy.Actor); ok {
		return actor, true
	}
	return policy.Actor{}, false
}
