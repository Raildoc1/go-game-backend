package grpcutils

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(serverAddress string) (*grpc.ClientConn, error) {
	options := grpc.WithTransportCredentials(insecure.NewCredentials())
	return grpc.NewClient(serverAddress, options)
}
