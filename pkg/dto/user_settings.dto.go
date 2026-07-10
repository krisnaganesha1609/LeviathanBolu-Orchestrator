package dto

import "github.com/krisnaganesha1609/LeviathanBolu-BE/pkg/domain"

type UpdateUserSettingsRequest struct {
	AssistantName string                  `json:"assistant_name"`
	WakeWords     []domain.WakeWordConfig `json:"wake_words" validate:"omitempty,dive,required,min=2"`
	Language      string                  `json:"language"`

	PreferredLLM string `json:"preferred_llm"`
	PreferredTTS string `json:"preferred_tts"`

	Theme string `json:"theme"`
}

type UserSettingsResponse struct {
	AssistantName string                  `json:"assistant_name"`
	WakeWords     []domain.WakeWordConfig `json:"wake_words"`
	Language      string                  `json:"language"`
	PreferredLLM  string                  `json:"preferred_llm"`
	PreferredTTS  string                  `json:"preferred_tts"`
	Theme         string                  `json:"theme"`
}
