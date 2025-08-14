package models

import (
	playerstoragepb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/protoutils"

	"github.com/google/uuid"
)

func FindUserByLoginTokenRequestToProto(loginToken uuid.UUID) *playerstoragepb.FindUserByLoginTokenRequest {
	return &playerstoragepb.FindUserByLoginTokenRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

func FindUserByLoginTokenResponseFromProto(resp *playerstoragepb.FindUserByLoginTokenResponse) (userID int64) {
	return *resp.UserID
}
