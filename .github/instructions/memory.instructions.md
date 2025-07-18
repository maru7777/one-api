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

## Claude Messages API: Universal Implementation

- **Architecture Decision:** Implemented universal Claude Messages API support across ALL adaptors using a conversion-based approach:
  ```
  Claude Messages Request → OpenAI Format → Native Adaptor Format
  ```
- **Key Implementation:**
  - Added `ConvertClaudeRequest()` method to `adaptor.Adaptor` interface
  - All 12+ adaptors implement Claude Messages conversion (Anthropic native, others via OpenAI intermediate)
  - Standardized context keys in `common/ctxkey/key.go`: `ClaudeMessagesConversion`, `OriginalClaudeRequest`
  - New endpoint: `POST /v1/messages` with full middleware integration

- **Billing Parity:** Complete billing feature parity with Chat Completion API:
  - **Two-phase billing:** Pre-consumption quota reservation + post-consumption accurate billing
  - **Four-layer pricing:** Channel overrides → Adapter defaults → Global pricing → Final fallback
  - **Comprehensive token counting:** Text (via OpenAI tokenization) + Images (Claude formula: `tokens = (width × height) / 750`) + Tools (schema tokenization)
  - **Special features:** Structured output detection (25% surcharge), multimodal content, tool calling costs
  - **Centralized billing:** Uses `billing.PostConsumeQuotaDetailed()` for consistency

- **Critical Subtleties:**
  - **Image token calculation:** Uses Claude-specific formula, handles base64/URL/file sources with reasonable estimates
  - **Context marking:** Essential for response conversion - adaptors mark requests for proper response handling
  - **Token counting fallback:** Uses character-based estimation when tiktoken fails (4 chars/token ratio)
  - **Universal compatibility:** Works with ANY channel type through adaptor conversion layer

- **Testing Strategy:**
  - Comprehensive unit tests for request validation, token counting, relay mode detection
  - MockAdaptor updated with `ConvertClaudeRequest()` method for test compatibility
  - All existing tests maintained - no breaking changes

## General Project Practices

- **Error Handling:**
  - Always use `github.com/Laisky/errors/v2` for error wrapping; never return bare errors.
- **Context Keys:**
  - ALL context keys must be pre-defined in `common/ctxkey/key.go` for consistency and maintainability
- **Package Management:**
  - Always use appropriate package managers (npm, pip, cargo, etc.) instead of manually editing package files
- **Testing:**
  - All bug fixes and features must be covered by unit tests. No temporary scripts.
- **Time Handling:**
  - Always use UTC for server, DB, and API time.
- **Golang ORM:**
  - Use `gorm.io/gorm` for writes; prefer SQL for reads to minimize DB load.

## Handover Guidance
- **Claude Messages API:** Fully production-ready with complete billing parity. See `docs/arch/api_billing.md` for comprehensive documentation.
- **Billing Architecture:** The three-layer system is now four-layer with channel-specific overrides as highest priority.
- **Adaptor Pattern:** All new API formats should follow the Claude Messages pattern: interface method + universal conversion + context marking.
- **Critical Files:**
  - `relay/controller/claude_messages.go` - Main implementation
  - `relay/adaptor/interface.go` - Interface definition
  - `common/ctxkey/key.go` - Context key definitions
  - `docs/arch/api_billing.md` - Complete billing documentation
- See `.github/memory.prompt` for detailed development context and architectural decisions.
