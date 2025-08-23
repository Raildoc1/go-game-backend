// Package futils contains helper types and utilities for working with
// context-aware functions.
package futils

import "context"

// CtxF represents a function that accepts a context and returns an error.
// It is commonly used for executing context-aware logic blocks.
type CtxF func(ctx context.Context) error
