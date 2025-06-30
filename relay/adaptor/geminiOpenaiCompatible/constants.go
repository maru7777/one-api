package geminiOpenaiCompatible

import "strings"

// https://ai.google.dev/models/gemini

var ModelList = []string{
	"gemini-2.5-pro",
	"gemini-2.5-flash",
	"gemini-2.0-flash-lite",
	"gemini-2.0-flash",
	"gemini-1.5-pro",
	"gemini-1.5-flash",
}

const (
	ModalityText  = "TEXT"
	ModalityImage = "IMAGE"
)

// GetModelModalities returns the modalities of the model.
func GetModelModalities(model string) []string {
	if strings.Contains(model, "-image-generation") {
		return []string{ModalityText, ModalityImage}
	}

	// Until 2025-03-26, the following models do not accept the responseModalities field
	if model == "aqa" ||
		strings.HasPrefix(model, "gemini-2.5") ||
		strings.HasPrefix(model, "gemma") ||
		strings.HasPrefix(model, "text-embed") {
		return nil
	}

	return []string{ModalityText}
}
