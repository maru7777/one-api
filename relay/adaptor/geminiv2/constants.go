package geminiv2

import "strings"

// https://ai.google.dev/models/gemini

var ModelList = []string{
	"gemini-pro", "gemini-1.0-pro",
	"gemma-2-2b-it", "gemma-2-9b-it", "gemma-2-27b-it",
	"gemma-3-27b-it",
	"gemini-1.5-flash", "gemini-1.5-flash-8b",
	"gemini-1.5-pro", "gemini-1.5-pro-experimental",
	"text-embedding-004", "aqa",
	"gemini-2.0-flash", "gemini-2.0-flash-exp",
	"gemini-2.0-flash-lite",
	"gemini-2.0-flash-thinking-exp-01-21",
	"gemini-2.0-flash-exp-image-generation",
	"gemini-2.5-flash-preview-04-17", "gemini-2.5-flash-preview-05-20",
	"gemini-2.0-pro-exp-02-05",
	"gemini-2.5-pro-exp-03-25", "gemini-2.5-pro-preview-05-06", "gemini-2.5-pro-preview-06-05",
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
