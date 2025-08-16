package models

import (
	playerstoragepb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/protoutils"

	"github.com/google/uuid"
)

// AddUserRequestToProto converts the login token to a protobuf AddUserRequest.
func AddUserRequestToProto(loginToken uuid.UUID) *playerstoragepb.AddUserRequest {
	return &playerstoragepb.AddUserRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

// AddUserResponseFromProto extracts the user ID from the protobuf response.
func AddUserResponseFromProto(resp *playerstoragepb.AddUserResponse) (userID int64) {
	return *resp.UserID
}
