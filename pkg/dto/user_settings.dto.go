package dto

type UpdateUserSettingsRequest struct {
	AssistantName string   `json:"assistant_name"`
	WakeWord      []string `json:"wake_word" validate:"dive"`
	Language      string   `json:"language"`

	PreferredLLM string `json:"preferred_llm"`
	PreferredTTS string `json:"preferred_tts"`

	Theme string `json:"theme"`
}

type UserSettingsResponse struct {
	AssistantName string   `json:"assistant_name"`
	WakeWord      []string `json:"wake_word"`
	Language      string   `json:"language"`
	PreferredLLM  string   `json:"preferred_llm"`
	PreferredTTS  string   `json:"preferred_tts"`
	Theme         string   `json:"theme"`
}
