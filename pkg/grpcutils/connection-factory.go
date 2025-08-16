package grpcutils

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Connect dials the gRPC server at the provided address using insecure
// transport credentials and returns the established connection.
func Connect(serverAddress string) (*grpc.ClientConn, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	return grpc.NewClient(serverAddress, options)
}
