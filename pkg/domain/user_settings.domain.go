package domain

import "github.com/google/uuid"

type UserSettings struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`

	AssistantName string
	// WakeWord is stored as a JSON array column. Plain []string has no
	// native Postgres mapping in GORM/pgx, so without this tag inserts and
	// scans on this field fail at runtime.
	WakeWord []string `gorm:"serializer:json;type:jsonb"`
	Language string

	PreferredLLM string
	PreferredTTS string

	Theme string

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
