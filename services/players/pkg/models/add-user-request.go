package models

import (
	"github.com/google/uuid"
	playerstoragepb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/protoutils"
)

func AddUserRequestToProto(loginToken uuid.UUID) *playerstoragepb.AddUserRequest {
	return &playerstoragepb.AddUserRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

func AddUserResponseFromProto(resp *playerstoragepb.AddUserResponse) (userID int64) {
	return *resp.UserID
}
