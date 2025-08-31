// Package postgresstore provides transaction context helpers for PostgreSQL.
package postgresstore

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type ctxField string

const txCtxField ctxField = "tx"

func ctxWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txCtxField, tx)
}

func txFromCtx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txCtxField).(pgx.Tx); ok {
		return tx
	}
	return nil
}
