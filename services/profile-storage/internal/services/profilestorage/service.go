package profilestorage

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines operations for accessing player credentials.
type Repository interface {
	AddUser(ctx context.Context, loginToken uuid.UUID) (int64, error)
	FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (int64, error)
}

// Service provides profile storage operations.
type Service struct {
	repo Repository
}

// New creates a new Service.
func New(repo Repository) *Service {
	return &Service{repo: repo}
}

// AddUser creates a new user and returns its ID.
func (s *Service) AddUser(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	return s.repo.AddUser(ctx, loginToken)
}

// FindUserByLoginToken looks up a user ID by login token.
func (s *Service) FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (int64, error) {
	return s.repo.FindUserByLoginToken(ctx, loginToken)
}
