package models

import (
	playerstoragepb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/protoutils"

	"github.com/google/uuid"
)

func AddUserRequestToProto(loginToken uuid.UUID) *playerstoragepb.AddUserRequest {
	return &playerstoragepb.AddUserRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

func AddUserResponseFromProto(resp *playerstoragepb.AddUserResponse) (userID int64) {
	return *resp.UserID
}
