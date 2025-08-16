package protoutils

import (
	"fmt"
	"go-game-backend/gen/dto"

	"github.com/google/uuid"
)

// UUIDFromProto converts a protobuf UUID message into a uuid.UUID value.
func UUIDFromProto(pbUUID *dto.UUID) (uuid.UUID, error) {
	res, err := uuid.FromBytes(pbUUID.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("uuid.FromBytes failed: %w", err)
	}
	return res, nil
}

// UUIDToProto converts a uuid.UUID into its protobuf representation.
func UUIDToProto(uuid uuid.UUID) *dto.UUID {
	return &dto.UUID{
		Value: uuid[:],
	}
}
