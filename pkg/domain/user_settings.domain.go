package domain

import "github.com/google/uuid"

// WakeWordConfig memetakan satu wake word ke personality yang diaktifkannya.
type WakeWordConfig struct {
	Word        string `json:"word"`
	Personality string `json:"personality"` // "bolu" | "leviathan"
}

type UserSettings struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`

	AssistantName string
	WakeWords     []WakeWordConfig `gorm:"serializer:json;type:jsonb;column:wake_words"`
	Language      string

	PreferredLLM string
	PreferredTTS string

	Theme string

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
