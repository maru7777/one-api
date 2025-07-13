---
applyTo: "**/*"
---

# Project Memory & Handover Instructions

## Gemini Adapter: Function Schema Cleaning

- **Problem:** Gemini API rejects OpenAI-style function schemas containing unsupported fields like `additionalProperties`, `description`, and `strict` (especially in nested objects), causing 400 errors.
- **Solution:**
  - Implemented a recursive cleaning function (`cleanFunctionParameters`) that:
    - Removes `additionalProperties` at all levels.
    - Removes `description` and `strict` only at the top level (preserved in nested objects).
    - Preserves all other fields required for Gemini compatibility.
  - Updated the request conversion logic to use this cleaning function for both `Tools` and `Functions`.
  - Added robust type assertions and fallback logic to ensure type safety.
- **Testing:**
  - Comprehensive unit tests were added to verify:
    - Recursive removal of `additionalProperties`.
    - Top-level-only removal of `description` and `strict`.
    - Preservation of all required fields and structure.
    - Compatibility with the exact schema that caused the original error.
  - All tests pass, including a regression test for the original log error.
- **Subtlety:**
  - Only remove `description` and `strict` at the top level; nested objects may require these fields for correct schema semantics.
  - Always type assert cleaned parameters to `map[string]any` before assignment to `model.Function.Parameters`.

## General Project Practices

- **Error Handling:**
  - Always use `github.com/Laisky/errors/v2` for error wrapping; never return bare errors.
- **Testing:**
  - All bug fixes and features must be covered by unit tests. No temporary scripts.
- **Time Handling:**
  - Always use UTC for server, DB, and API time.
- **Golang ORM:**
  - Use `gorm.io/gorm` for writes; prefer SQL for reads to minimize DB load.

## Handover Guidance
- Review the latest tests in `relay/adaptor/gemini/main_test.go` for regression coverage.
- The cleaning logic is critical for Gemini compatibility and should be maintained if Gemini API changes its schema requirements.
- See `.github/memory.prompt` for a summary of recent development context and decisions.
