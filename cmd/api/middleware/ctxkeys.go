package middleware

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ctxKey int

const (
	ctxKeyBody ctxKey = iota + 1
	ctxKeyDb
	ctxKeyLog
)

func WithBody(ctx context.Context, body []byte) context.Context {
	return context.WithValue(ctx, ctxKeyBody, body)
}

func BodyFromContext(ctx context.Context) ([]byte, bool) {
	body, ok := ctx.Value(ctxKeyBody).([]byte)
	return body, ok
}

func WithDB(ctx context.Context, db *pgxpool.Pool) context.Context {
	return context.WithValue(ctx, ctxKeyDb, db)
}

func DBFromContext(ctx context.Context) (*pgxpool.Pool, bool) {
	db, ok := ctx.Value(ctxKeyDb).(*pgxpool.Pool)
	return db, ok
}

func WithLogger(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxKeyLog, log)
}

func LoggerFromContext(ctx context.Context) (*zap.SugaredLogger, bool) {
	log, ok := ctx.Value(ctxKeyLog).(*zap.SugaredLogger)
	return log, ok
}
