package models

import (
	profilestoragepb "go-game-backend/gen/profilestorage"
	"go-game-backend/pkg/protoutils"

	"github.com/google/uuid"
)

// FindUserByLoginTokenRequestToProto converts a login token into the protobuf
// request type used to find a user.
func FindUserByLoginTokenRequestToProto(loginToken uuid.UUID) *profilestoragepb.FindUserByLoginTokenRequest {
	return &profilestoragepb.FindUserByLoginTokenRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

// FindUserByLoginTokenResponseFromProto extracts the user ID from the protobuf
// response.
func FindUserByLoginTokenResponseFromProto(resp *profilestoragepb.FindUserByLoginTokenResponse) (userID int64) {
	return *resp.UserID
}
