package domain

type Prompt struct {
	Model

	UserID       ID
	PromptText   string
	SettingsHash string // Hash of the generation settings used to create this prompt
}

func NewPrompt(userID ID, promptText, settingsHash string) Prompt {
	return Prompt{
		Model:        NewModel(),
		UserID:       userID,
		PromptText:   promptText,
		SettingsHash: settingsHash,
	}
}
