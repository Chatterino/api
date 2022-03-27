package livestreamfails

import (
	"context"
	"errors"
)

type contextKey string

var (
	contextClipID = contextKey("clipID")

	errMissingClipID = errors.New("missing clip ID in context")
)

func contextWithClipID(ctx context.Context, clipID string) context.Context {
	ctx = context.WithValue(ctx, contextClipID, clipID)
	return ctx
}

func clipIDFromContext(ctx context.Context) (string, error) {
	clipID, ok := ctx.Value(contextClipID).(string)
	if !ok {
		return "", errMissingClipID
	}

	return clipID, nil
}
