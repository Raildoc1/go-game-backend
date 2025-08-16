package models

import (
	playerstoragepb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/protoutils"

	"github.com/google/uuid"
)

// FindUserByLoginTokenRequestToProto converts a login token into the protobuf
// request type used to find a user.
func FindUserByLoginTokenRequestToProto(loginToken uuid.UUID) *playerstoragepb.FindUserByLoginTokenRequest {
	return &playerstoragepb.FindUserByLoginTokenRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

// FindUserByLoginTokenResponseFromProto extracts the user ID from the protobuf
// response.
func FindUserByLoginTokenResponseFromProto(resp *playerstoragepb.FindUserByLoginTokenResponse) (userID int64) {
	return *resp.UserID
}
