package protoutils

import (
	"fmt"
	"go-game-backend/gen/dto"

	"github.com/google/uuid"
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
