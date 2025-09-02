package postgresstore

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// txQuerier is a generic interface for sqlc query structs that can bind to a transaction.
type txQuerier[Q any] interface {
	WithTx(pgx.Tx) Q
}

// BaseRepo provides helper methods to access queries bound to context transactions.
type BaseRepo[Q txQuerier[Q]] struct {
	queries Q
}

// NewBaseRepo creates a BaseRepo using the provided queries instance.
func NewBaseRepo[Q txQuerier[Q]](queries Q) BaseRepo[Q] {
	return BaseRepo[Q]{queries: queries}
}

// Q returns queries bound to the transaction from context if present.
func (r BaseRepo[Q]) Q(ctx context.Context) Q {
	if tx := txFromCtx(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}
