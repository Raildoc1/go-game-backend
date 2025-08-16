package grpcserver

import (
	"context"
	"go-game-backend/pkg/protoutils"

	pb "go-game-backend/gen/profilestorage"

	logic "go-game-backend/services/profile-storage/internal/services/profilestorage"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedProfileStorageServiceServer
	svc *logic.Service
}

func NewServer(svc *logic.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) AddUser(ctx context.Context, req *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	loginToken, err := protoutils.UUIDFromProto(req.LoginToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid login token: %v", err)
	}
	id, err := s.svc.AddUser(ctx, loginToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "add user failed: %v", err)
	}
	return &pb.AddUserResponse{UserID: &id}, nil
}

func (s *Server) FindUser(ctx context.Context, req *pb.FindUserByLoginTokenRequest) (*pb.FindUserByLoginTokenResponse, error) {
	loginToken, err := protoutils.UUIDFromProto(req.LoginToken)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid login token: %v", err)
	}
	id, err := s.svc.FindUserByLoginToken(ctx, loginToken)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "find user failed: %v", err)
	}
	return &pb.FindUserByLoginTokenResponse{UserID: &id}, nil
}
