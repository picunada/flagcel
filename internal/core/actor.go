package core

import "context"

// Actor identifies who performed a mutation. It is carried on the request
// context by the admin auth middleware and read by the store when writing
// audit metadata. When auth is disabled the actor is nil.
type Actor struct {
	ID    string
	Label string
}

type actorContextKey struct{}

// WithActor returns a copy of ctx carrying the given actor.
func WithActor(ctx context.Context, actor *Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

// ActorFromContext returns the actor carried on ctx, or nil if none is set.
func ActorFromContext(ctx context.Context) *Actor {
	actor, _ := ctx.Value(actorContextKey{}).(*Actor)
	return actor
}
