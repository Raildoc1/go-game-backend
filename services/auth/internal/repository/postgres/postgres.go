package postgresrepo

import (
	"context"
	auth "go-game-backend/services/auth/internal/services/auth"
)

type Repository struct{}

var _ auth.UserRepository = (*Repository)(nil)

func New() *Repository {
	return &Repository{}
}

func (r Repository) AddUser(ctx context.Context, loginToken string) (userID int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (r Repository) FindUserByLoginToken(ctx context.Context, loginToken string) (userID int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (r Repository) FindUserByRefreshToken(ctx context.Context, loginToken string) (userID int64, err error) {
	//TODO implement me
	panic("implement me")
}
