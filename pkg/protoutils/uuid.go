package protoutils

import (
	"fmt"
	"github.com/google/uuid"
	"go-game-backend/gen/dto"
)

func UUIDFromProto(pbUUID *dto.UUID) (uuid.UUID, error) {
	res, err := uuid.FromBytes(pbUUID.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.FromBytes failed: %w", err)
	}
	return res, nil
}

func UUIDToProto(uuid uuid.UUID) *dto.UUID {
	return &dto.UUID{
		Value: uuid[:],
	}
}
