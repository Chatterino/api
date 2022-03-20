package wikipedia

import (
	"context"
	"errors"
)

type contextKey string

var (
	contextLocaleCode = contextKey("localeCode")
	contextArticleID  = contextKey("articleID")

	errMissingArticleValues = errors.New("missing article values in context")
)

func contextWithArticleValues(ctx context.Context, localeCode, articleID string) context.Context {
	ctx = context.WithValue(ctx, contextLocaleCode, localeCode)
	ctx = context.WithValue(ctx, contextArticleID, articleID)
	return ctx
}

func articleValuesFromContext(ctx context.Context) (string, string, error) {
	articleID, ok := ctx.Value(contextArticleID).(string)
	if !ok {
		return "", "", errMissingArticleValues
	}

	localeCode, ok := ctx.Value(contextLocaleCode).(string)
	if !ok {
		return "", "", errMissingArticleValues
	}

	return localeCode, articleID, nil
}
