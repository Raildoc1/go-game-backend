// Package futils contains helper types and utilities for working with
// context-aware functions.
package futils

import "context"

// CtxF represents a function that accepts a context and returns an error.
// It is commonly used for executing context-aware logic blocks.
type CtxF func(ctx context.Context) error

// CtxFT represents a function that accepts a context and extra argument and returns an error.
// It is commonly used for executing context-aware logic blocks.
type CtxFT[T any] func(ctx context.Context, data T) error
