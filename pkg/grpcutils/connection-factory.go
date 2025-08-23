// Package grpcutils provides helpers for establishing gRPC connections and
// working with gRPC clients.
package grpcutils

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Connect dials the gRPC server at the provided address using insecure
// transport credentials and returns the established connection.
func Connect(serverAddress string) (*grpc.ClientConn, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(serverAddress, options)
	if err != nil {
		return nil, fmt.Errorf("could not create grpc connection: %w", err)
	}
	return conn, nil
}
