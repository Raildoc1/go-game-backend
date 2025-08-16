package models

import (
	profilestoragepb "go-game-backend/gen/profilestorage"
	"go-game-backend/pkg/protoutils"

	"github.com/google/uuid"
)

// AddUserRequestToProto converts the login token to a protobuf AddUserRequest.
func AddUserRequestToProto(loginToken uuid.UUID) *profilestoragepb.AddUserRequest {
	return &profilestoragepb.AddUserRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

// AddUserResponseFromProto extracts the user ID from the protobuf response.
func AddUserResponseFromProto(resp *profilestoragepb.AddUserResponse) (userID int64) {
	return *resp.UserID
}
