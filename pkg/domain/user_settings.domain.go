package domain

import "github.com/google/uuid"

type UserSettings struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`

	AssistantName string
	WakeWord      []string
	Language      string

	PreferredLLM string
	PreferredTTS string

	Theme string

	User User `gorm:"foreignKey:UserID"`
}
