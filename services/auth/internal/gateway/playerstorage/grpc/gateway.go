package grpc

import (
	"context"
	"fmt"
	"go-game-backend/pkg/grpcutils"
	"go-game-backend/services/players/pkg/models"

	"github.com/google/uuid"
	pb "go-game-backend/gen/playerstorage"

	auth "go-game-backend/services/auth/internal/services/auth"

	"google.golang.org/grpc"
)

// Config holds settings for connecting to the player storage gRPC service.
type Config struct {
	ServerAddress string `yaml:"server-address"`
}

// Gateway provides access to the remote player storage service via gRPC.
type Gateway struct {
	conn          *grpc.ClientConn
	playerStorage pb.PlayerStorageServiceClient
}

var _ auth.PlayerStorageGateway = (*Gateway)(nil)

// New creates a new Gateway using the given configuration.
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

// Stop closes the underlying gRPC connection.
func (g *Gateway) Stop() error {
	err := g.conn.Close()
	if err != nil {
		return fmt.Errorf("error closing gRPC connection: %w", err)
	}
	return nil
}

// AddUser creates a new user in the player storage service using the provided
// login token and returns the new user ID.
func (g *Gateway) AddUser(ctx context.Context, loginToken uuid.UUID) (userID int64, err error) {
	req := models.AddUserRequestToProto(loginToken)
	resp, err := g.playerStorage.AddUser(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to add user: %w", err)
	}
	return models.AddUserResponseFromProto(resp), nil
}

// FindUserByLoginToken retrieves a user ID associated with the specified login
// token.
func (g *Gateway) FindUserByLoginToken(ctx context.Context, loginToken uuid.UUID) (userID int64, err error) {
	req := models.FindUserByLoginTokenRequestToProto(loginToken)
	resp, err := g.playerStorage.FindUser(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to find user: %w", err)
	}
	return models.FindUserByLoginTokenResponseFromProto(resp), nil
}
