package utils

import (
	"context"

	"github.com/google/uuid"
)

const traceIDKey = "traceID"

func NewContextWithTraceID() context.Context {
	traceID := uuid.New().String()
	return context.WithValue(context.Background(), traceIDKey, traceID)
}
func NewContext(traceID string) context.Context {
	return context.WithValue(context.Background(), traceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return "unknown"
}
