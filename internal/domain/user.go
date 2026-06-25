package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	OAuthUserID  uuid.UUID
	EncryptedDEK []byte
	CreatedAt    time.Time
}

type Datanode struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	Host      string
	Port      int
	SSHUser   string
	LastJobID *string
	CreatedAt time.Time
}

type VM struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	DatanodeID           uuid.UUID
	DatanodeName         string
	Name                 string
	Host                 string
	Port                 int
	EncryptedCredentials []byte
	CreatedAt            time.Time
}
