package lib

import (
	"context"

	"github.com/Bendomey/fincore-engine/internal/models"
)

type contextKey string

const clientContextKey contextKey = "fin-core-client"

func WithClient(ctx context.Context, client *models.Client) context.Context {
	return context.WithValue(ctx, clientContextKey, client)
}

func ClientFromContext(ctx context.Context) (*models.Client, bool) {
	client, ok := ctx.Value(clientContextKey).(*models.Client)
	return client, ok
}
