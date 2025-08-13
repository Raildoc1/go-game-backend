package models

import (
	"github.com/google/uuid"
	playerstoragepb "go-game-backend/gen/playerstorage"
	"go-game-backend/pkg/protoutils"
)

func FindUserByLoginTokenRequestToProto(loginToken uuid.UUID) *playerstoragepb.FindUserByLoginTokenRequest {
	return &playerstoragepb.FindUserByLoginTokenRequest{
		LoginToken: protoutils.UUIDToProto(loginToken),
	}
}

func FindUserByLoginTokenResponseFromProto(resp *playerstoragepb.FindUserByLoginTokenResponse) (userID int64) {
	return *resp.UserID
}
