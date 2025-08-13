package grpc

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	pb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/grpcutils"
	auth "go-game-backend/services/auth/internal/services/auth"
	"go-game-backend/services/players/pkg/models"
	"google.golang.org/grpc"
)

type Config struct {
	ServerAddress string `yaml:"server-address"`
}

type Gateway struct {
	conn          *grpc.ClientConn
	playerStorage pb.PlayerStorageServiceClient
}

var _ auth.PlayerStorageGateway = (*Gateway)(nil)

func New(cfg *Config) (*Gateway, error) {
	conn, err := grpcutils.Connect(cfg.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	pb.NewPlayerStorageServiceClient(conn)
	return &Gateway{
		conn:          conn,
		playerStorage: pb.NewPlayerStorageServiceClient(conn),
	}, nil
}

func (g *Gateway) Stop() error {
	err := g.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing gRPC connection: %w", err)
	}
	return nil
}

func (g *Gateway) AddUser(ctx context.Context, loginToken uuid.UUID) (userID int64, err error) {
	req := models.AddUserRequestToProto(loginToken)
	resp, err := g.playerStorage.AddUser(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to add user: %w", err)
	}
	return models.AddUserResponseFromProto(resp), nil
}

func (g *Gateway) FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (userID int64, err error) {
	req := models.FindUserByLoginTokenRequestToProto(loginToken)
	resp, err := g.playerStorage.FindUser(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to find user: %w", err)
	}
	return models.FindUserByLoginTokenResponseFromProto(resp), nil
}
