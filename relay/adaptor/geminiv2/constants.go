package geminiv2

import "strings"

// https://ai.google.dev/models/gemini

var ModelList = []string{
	"gemini-pro", "gemini-1.0-pro",
	// "gemma-2-2b-it", "gemma-2-9b-it", "gemma-2-27b-it",
	"gemini-1.5-flash", "gemini-1.5-flash-8b",
	"gemini-1.5-pro", "gemini-1.5-pro-experimental",
	"text-embedding-004", "aqa",
	"gemini-2.0-flash", "gemini-2.0-flash-exp",
	"gemini-2.0-flash-lite",
	"gemini-2.0-flash-thinking-exp-01-21",
	"gemini-2.0-pro-exp-02-05",
	"gemini-2.0-flash-exp-image-generation",
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

	return []string{ModalityText}
}
