# API Billing

Refs:

- [Billing system](./billing.md)

- [API Billing](#api-billing)
  - [Chat Completion API](#chat-completion-api)
    - [Billing Architecture Overview](#billing-architecture-overview)
    - [Core Billing Components](#core-billing-components)
      - [1. Token-Based Billing Fields](#1-token-based-billing-fields)
      - [2. Additional Cost Components](#2-additional-cost-components)
    - [Billing Calculation Formula](#billing-calculation-formula)
      - [Detailed Breakdown](#detailed-breakdown)
    - [Pricing Resolution System](#pricing-resolution-system)
      - [Layer 1: Channel-Specific Overrides (Highest Priority)](#layer-1-channel-specific-overrides-highest-priority)
      - [Layer 2: Adapter Default Pricing (Second Priority)](#layer-2-adapter-default-pricing-second-priority)
      - [Layer 3: Global Pricing Fallback (Third Priority)](#layer-3-global-pricing-fallback-third-priority)
      - [Layer 4: Final Default (Lowest Priority)](#layer-4-final-default-lowest-priority)
    - [Token Counting Methods](#token-counting-methods)
      - [Text Content](#text-content)
      - [Image Content](#image-content)
      - [Audio Content](#audio-content)
    - [Special Billing Features](#special-billing-features)
      - [1. Structured Output Billing](#1-structured-output-billing)
      - [2. Function Calling Billing](#2-function-calling-billing)
      - [3. Reasoning Model Billing](#3-reasoning-model-billing)
      - [4. Multi-modal Content Billing](#4-multi-modal-content-billing)
    - [Pre-consumption vs Post-consumption](#pre-consumption-vs-post-consumption)
      - [Pre-consumption Phase](#pre-consumption-phase)
      - [Post-consumption Phase](#post-consumption-phase)
    - [Error Handling and Refunds](#error-handling-and-refunds)
      - [Automatic Refund Scenarios](#automatic-refund-scenarios)
      - [Refund Implementation](#refund-implementation)
    - [Usage Logging and Tracking](#usage-logging-and-tracking)
      - [Log Structure](#log-structure)
      - [Billing Metrics Tracked](#billing-metrics-tracked)
    - [Implementation Files](#implementation-files)
      - [Core Billing Logic](#core-billing-logic)
      - [Token Counting](#token-counting)
      - [Structured Output](#structured-output)
      - [Model Pricing](#model-pricing)
  - [OpenAI Response API](#openai-response-api)
    - [Implementation Architecture](#implementation-architecture)
      - [Request Processing Flow](#request-processing-flow)
      - [Dual API Support](#dual-api-support)
    - [Core Implementation Components](#core-implementation-components)
      - [Relay Mode and Routing](#relay-mode-and-routing)
      - [Response API Controller](#response-api-controller)
      - [OpenAI Adapter Enhancements](#openai-adapter-enhancements)
    - [Billing Integration](#billing-integration)
      - [Token Counting and Estimation](#token-counting-and-estimation)
      - [Pre-consumption and Post-consumption](#pre-consumption-and-post-consumption)
      - [Pricing Compatibility](#pricing-compatibility)
    - [Advanced Features Support](#advanced-features-support)
      - [Streaming Implementation](#streaming-implementation)
      - [Structured Output Support](#structured-output-support)
      - [Function Calling Support](#function-calling-support)
    - [API Endpoints and Usage](#api-endpoints-and-usage)
    - [Error Handling and Validation](#error-handling-and-validation)
    - [Implementation Benefits](#implementation-benefits)


## Chat Completion API

The Chat Completion API implements a sophisticated billing system that accurately calculates costs based on multiple factors including token usage, model pricing, completion ratios, and additional features like structured output and function calling.

### Billing Architecture Overview

The Chat Completion API billing follows a **two-phase approach**:

1. **Pre-consumption Phase**: Reserve quota based on estimated prompt tokens
2. **Post-consumption Phase**: Calculate final billing based on actual usage and refund/charge the difference

```mermaid
sequenceDiagram
    participant Client
    participant Controller
    participant PricingManager
    participant QuotaManager
    participant Adapter
    participant Database

    Client->>Controller: Chat Completion Request
    Controller->>PricingManager: Get Model Pricing
    PricingManager-->>Controller: Model Ratio & Completion Ratio

    Controller->>QuotaManager: Pre-consume Quota (Prompt Tokens)
    QuotaManager->>Database: Reserve Quota
    Database-->>QuotaManager: Quota Reserved
    QuotaManager-->>Controller: Pre-consumption OK

    Controller->>Adapter: Process Request
    Adapter-->>Controller: Response + Usage Details

    Controller->>QuotaManager: Post-consume Quota (Final Billing)
    QuotaManager->>Database: Final Billing & Refund
    Database-->>QuotaManager: Billing Complete
    QuotaManager-->>Controller: Billing Complete

    Controller-->>Client: Chat Completion Response
```

### Core Billing Components

#### 1. Token-Based Billing Fields

The system tracks and bills for the following token types:

**Primary Token Fields:**

- **`PromptTokens`**: Input tokens from user messages, system prompts, and context
- **`CompletionTokens`**: Output tokens generated by the model
- **`TotalTokens`**: Sum of prompt and completion tokens (for reference)

**Detailed Token Breakdown:**

- **`PromptTokensDetails.TextTokens`**: Text-based prompt tokens
- **`PromptTokensDetails.AudioTokens`**: Audio input tokens (converted to text equivalent)
- **`PromptTokensDetails.ImageTokens`**: Image input tokens (based on image size and detail)
- **`PromptTokensDetails.CachedTokens`**: Cached prompt tokens (may have different pricing)
- **`CompletionTokensDetails.TextTokens`**: Text-based completion tokens
- **`CompletionTokensDetails.AudioTokens`**: Audio output tokens
- **`CompletionTokensDetails.ReasoningTokens`**: Reasoning tokens for reasoning models (like o1)

#### 2. Additional Cost Components

**`ToolsCost`**: Additional charges for special features:

- **Structured Output**: 25% surcharge on completion tokens when using `json_schema` response format
- **Function Calling**: Additional costs for tool usage (model-dependent)
- **Web Search**: Variable costs based on search context size (small/medium/large)

### Billing Calculation Formula

The final quota calculation follows this comprehensive formula:

```text
Final Quota = Base Token Cost + Tools Cost

Where:
Base Token Cost = (PromptTokens + CompletionTokens × CompletionRatio) × ModelRatio × GroupRatio
Tools Cost = Structured Output Cost + Function Call Cost + Web Search Cost
```

#### Detailed Breakdown

**1. Base Token Cost Calculation:**

```go
baseTokenCost = (promptTokens + completionTokens * completionRatio) * modelRatio * groupRatio
```

**2. Structured Output Cost (when applicable):**

```go
structuredOutputCost = math.Ceil(completionTokens * 0.25 * modelRatio)
```

**3. Final Quota:**

```go
finalQuota = baseTokenCost + toolsCost
if modelRatio != 0 && finalQuota <= 0 {
    finalQuota = 1  // Minimum charge
}
```

### Pricing Resolution System

The Chat Completion API uses the **Four-Layer Pricing Resolution** system:

#### Layer 1: Channel-Specific Overrides (Highest Priority)

Custom pricing set by administrators for specific channels:

```json
{
  "model_ratio": { "gpt-4": 0.03, "gpt-3.5-turbo": 0.002 },
  "completion_ratio": { "gpt-4": 3.0, "gpt-3.5-turbo": 1.0 }
}
```

#### Layer 2: Adapter Default Pricing (Second Priority)

Native pricing from channel adapters based on official provider pricing:

```go
// Example from OpenAI adapter
"gpt-4": {
    Ratio:           0.03 * MilliTokensUsd,    // $30 per 1M input tokens
    CompletionRatio: 2.0,                      // 2x multiplier for output tokens
}
```

#### Layer 3: Global Pricing Fallback (Third Priority)

Merged pricing from 13+ major adapters for comprehensive coverage:

- Automatically loaded from OpenAI, Anthropic, Gemini, Ali, Baidu, Zhipu, etc.
- Provides pricing for common models across different channels

#### Layer 4: Final Default (Lowest Priority)

Reasonable fallback pricing:

- **Model Ratio**: 2.5 USD per million tokens
- **Completion Ratio**: 1.0 (no multiplier)

### Token Counting Methods

#### Text Content

Uses tiktoken-based encoding specific to each model:

```go
// Message structure: <|start|>{role}\n{content}<|end|>\n
tokensPerMessage = 3  // For most models
tokensPerName = 1     // If name field is present
totalTokens = messageTokens + contentTokens + 3  // Reply primer
```

#### Image Content

Token calculation based on image dimensions and detail level:

```go
// Low detail: Fixed cost
lowDetailCost = 85 tokens

// High detail: Based on image tiles
numTiles = ceil(width/512) * ceil(height/512)
highDetailCost = numTiles * costPerTile + additionalCost

// Model-specific costs:
// GPT-4o: 170 tokens per tile + 85 base
// GPT-4o-mini: 2833 tokens per tile + 5667 base
```

#### Audio Content

Audio tokens converted to text-equivalent tokens:

```go
// Audio input tokens (per second)
audioPromptTokens = audioSeconds * audioTokensPerSecond * audioPromptRatio

// Audio output tokens
audioCompletionTokens = audioSeconds * audioTokensPerSecond * audioCompletionRatio

// Default rates:
audioTokensPerSecond = 10  // Conservative estimate
audioPromptRatio = 1.0
audioCompletionRatio = varies by model
```

### Special Billing Features

#### 1. Structured Output Billing

When using `response_format` with `json_schema`:

**Detection Logic:**

```go
if request.ResponseFormat != nil &&
   request.ResponseFormat.Type == "json_schema" &&
   request.ResponseFormat.JsonSchema != nil {
    // Apply 25% surcharge on completion tokens
    structuredOutputCost = ceil(completionTokens * 0.25 * modelRatio)
    usage.ToolsCost += structuredOutputCost
}
```

**Cost Impact:**

- Adds 25% surcharge to completion token costs
- Applied only to the completion tokens, not prompt tokens
- Calculated using the same model ratio as base pricing

#### 2. Function Calling Billing

Function calls are billed through the standard token mechanism:

- Function definitions count toward prompt tokens
- Function call responses count toward completion tokens
- No additional surcharge beyond standard token costs

#### 3. Reasoning Model Billing

For reasoning models (like o1-preview, o1-mini):

- **Reasoning tokens** are currently merged into completion tokens by most providers
- Future implementation may separate reasoning token billing
- Currently billed at standard completion token rates

#### 4. Multi-modal Content Billing

**Text + Image:**

```go
totalPromptTokens = textTokens + imageTokens
// Image tokens calculated based on resolution and detail level
```

**Text + Audio:**

```go
totalPromptTokens = textTokens + ceil(audioTokens * audioPromptRatio)
totalCompletionTokens = textTokens + ceil(audioTokens * audioCompletionRatio)
```

### Pre-consumption vs Post-consumption

#### Pre-consumption Phase

**Purpose**: Reserve quota to prevent over-spending
**Calculation**: Based on estimated prompt tokens only

```go
preConsumedQuota = promptTokens * modelRatio * groupRatio
```

**Token Estimation:**

- Counts all message content (text, images, audio)
- Includes system prompts and conversation history
- Does not include completion tokens (unknown at this stage)

#### Post-consumption Phase

**Purpose**: Final billing based on actual usage
**Calculation**: Complete formula with all components

```go
finalQuota = (promptTokens + completionTokens * completionRatio) * modelRatio * groupRatio + toolsCost
```

**Centralized Billing Implementation**: The ChatCompletion API now uses the centralized `billing.PostConsumeQuotaDetailed()` function in `relay/controller/helper.go:postConsumeQuota()` which ensures:

- Consistent billing logic with Response API
- Complete logging with all metadata fields (ElapsedTime, IsStream, SystemPromptReset)
- Proper token quota management via `model.PostConsumeTokenQuota()`
- Unified user and channel quota updates

**Quota Adjustment:**

- If `finalQuota > preConsumedQuota`: Charge additional amount
- If `finalQuota < preConsumedQuota`: Refund difference
- Minimum charge of 1 quota unit if model ratio > 0

### Error Handling and Refunds

#### Automatic Refund Scenarios

1. **Request Failure**: Full refund of pre-consumed quota
2. **Processing Error**: Full refund of pre-consumed quota
3. **Adapter Error**: Full refund of pre-consumed quota

#### Refund Implementation

```go
func ReturnPreConsumedQuota(ctx context.Context, quota int64, tokenId int) {
    if quota > 0 {
        err := PostConsumeTokenQuota(tokenId, -quota)  // Negative amount = refund
        if err != nil {
            logger.Error(ctx, "Failed to return pre-consumed quota: " + err.Error())
        }
    }
}
```

### Usage Logging and Tracking

#### Log Structure

```go
type Log struct {
    UserId:           int
    ChannelId:        int
    PromptTokens:     int     // Actual prompt tokens used
    CompletionTokens: int     // Actual completion tokens generated
    ModelName:        string  // Model used for the request
    TokenName:        string  // API token name
    Quota:            int     // Total quota consumed
    Content:          string  // Additional billing details
}
```

#### Billing Metrics Tracked

- **User-level**: Total quota used, request count
- **Channel-level**: Channel quota consumption
- **Model-level**: Per-model usage statistics
- **Token-level**: API token usage tracking

### Implementation Files

#### Core Billing Logic

- **`relay/controller/text.go`**: Main chat completion controller with pre-consumption
- **`relay/controller/helper.go`**: Post-consumption quota calculation
- **`relay/billing/billing.go`**: Core billing operations and logging
- **`relay/pricing/global.go`**: Four-layer pricing resolution system

#### Token Counting

- **`relay/adaptor/openai/token.go`**: Comprehensive token counting for all content types
- **`relay/controller/helper.go`**: Prompt token estimation

#### Structured Output

- **`relay/adaptor/openai/adaptor.go`**: Structured output cost calculation
- **`relay/adaptor/openai/structured_output_*_test.go`**: Test coverage for structured output billing

#### Model Pricing

- **`relay/adaptor/*/constants.go`**: Adapter-specific model pricing (25+ adapters)
- **`relay/billing/ratio/model.go`**: Legacy pricing compatibility functions

This comprehensive billing system ensures accurate cost calculation for all Chat Completion API features while maintaining backward compatibility and providing transparent pricing across all supported models and providers.

## OpenAI Response API

The OpenAI Response API represents a new interface for generating model responses with enhanced capabilities including stateful conversations, built-in tools, and advanced features. This section documents the **completed implementation** of direct Response API support with full billing integration.

### Implementation Status: ✅ COMPLETE

**Phase 1 Implementation Completed:**

- ✅ **Direct Response API Support**: Users can send requests directly in Response API format to `/v1/responses`
- ✅ **Native Response API Billing**: Dedicated billing path using the same pricing system as ChatCompletion
- ✅ **Streaming Support**: Full streaming compatibility with direct pass-through of Response API events
- ✅ **Backward Compatibility**: No breaking changes to existing ChatCompletion functionality
- ✅ **Feature Parity**: Structured output, function calling, and multi-modal content support

**Dual API Support:**

The system now supports both API formats simultaneously:

- **ChatCompletion API** (`/v1/chat/completions`): Automatic conversion to Response API upstream, with response conversion back to ChatCompletion format
- **Response API** (`/v1/responses`): Direct pass-through without conversion overhead, native Response API responses

### Implementation Architecture

#### Request Processing Flow

The system supports both ChatCompletion and direct Response API requests through a unified architecture:

```mermaid
sequenceDiagram
    participant User
    participant Controller
    participant OpenAIAdapter
    participant OpenAIAPI
    participant BillingSystem

    alt ChatCompletion Request (Existing)
        User->>Controller: ChatCompletion Request
        Controller->>OpenAIAdapter: Process Request
        OpenAIAdapter->>OpenAIAdapter: Convert to Response API
        OpenAIAdapter->>OpenAIAPI: Response API Request
        OpenAIAPI-->>OpenAIAdapter: Response API Response
        OpenAIAdapter->>OpenAIAdapter: Convert to ChatCompletion
        OpenAIAdapter-->>Controller: ChatCompletion Response
        Controller->>BillingSystem: Bill using ChatCompletion format
        Controller-->>User: ChatCompletion Response
    else Response API Request (New)
        User->>Controller: Response API Request
        Controller->>OpenAIAdapter: Process Request (Direct Pass-through)
        OpenAIAdapter->>OpenAIAPI: Response API Request (Direct)
        OpenAIAPI-->>OpenAIAdapter: Response API Response
        OpenAIAdapter-->>Controller: Response API Response (Direct)
        Controller->>BillingSystem: Bill using Response API format
        Controller-->>User: Response API Response
    end
```

#### Dual API Support

The implementation provides seamless support for both API formats:

**ChatCompletion API (Existing):**

- Requests sent to `/v1/chat/completions`
- Automatic conversion to Response API for upstream processing
- Response conversion back to ChatCompletion format
- Full backward compatibility maintained

**Response API (New):**

- Requests sent to `/v1/responses`
- Direct pass-through without conversion overhead
- Native Response API response format
- Full feature parity with OpenAI's Response API

### Core Implementation Components

#### Relay Mode and Routing

**Path Detection:** The system detects Response API requests by checking for `/v1/responses` path prefix in `relay/relaymode/helper.go:GetByPath()`

**Routing:** New endpoints added in `router/relay.go`:

- `POST /v1/responses` - Main Response API endpoint
- `GET /v1/responses/:response_id` - Get response (placeholder)
- `DELETE /v1/responses/:response_id` - Delete response (placeholder)
- `POST /v1/responses/:response_id/cancel` - Cancel response (placeholder)

**Controller Dispatch:** The main relay controller in `controller/relay.go:relayHelper()` routes Response API requests to `RelayResponseAPIHelper()`

#### Response API Controller

**Main Handler:** `RelayResponseAPIHelper()` in `relay/controller/response.go` handles Response API requests with full billing integration:

**Key Functions:**

- `getAndValidateResponseAPIRequest()` - Validates incoming Response API requests
- `getResponseAPIPromptTokens()` - Estimates input tokens for pre-consumption
- `preConsumeResponseAPIQuota()` - Reserves quota based on estimated tokens
- `postConsumeResponseAPIQuota()` - Calculates final billing and adjusts quota
- `getResponseAPIRequestBody()` - Prepares request for direct pass-through

**Channel Support:** Currently limited to OpenAI channels (channel type 1) only

**Pricing Integration:** Uses the same three-layer pricing system as ChatCompletion API via `pricing.GetModelRatioWithThreeLayers()`

#### OpenAI Adapter Enhancements

**Request Processing:** The `ConvertRequest()` method in `relay/adaptor/openai/adaptor.go` handles both API formats:

- **Response API requests:** Direct pass-through without conversion
- **ChatCompletion requests:** Automatic conversion to Response API format for upstream processing

**Response Processing:** The `DoResponse()` method routes responses based on relay mode:

- **Direct Response API:** Uses `ResponseAPIDirectHandler()` and `ResponseAPIDirectStreamHandler()` for pass-through
- **Converted ChatCompletion:** Uses `ResponseAPIHandler()` and `ResponseAPIStreamHandler()` for conversion back to ChatCompletion format

**Handler Functions in `relay/adaptor/openai/main.go`:**

- `ResponseAPIDirectHandler()` - Non-streaming Response API pass-through
- `ResponseAPIDirectStreamHandler()` - Streaming Response API pass-through
- `ResponseAPIHandler()` - Response API to ChatCompletion conversion
- `ResponseAPIStreamHandler()` - Streaming Response API to ChatCompletion conversion

### Billing Integration

#### Token Counting and Estimation

**Token Field Mapping:**

| Response API Field      | ChatCompletion Equivalent   | Billing Usage                  |
| ----------------------- | --------------------------- | ------------------------------ |
| `input_tokens`          | `prompt_tokens`             | Pre-consumption calculation    |
| `output_tokens`         | `completion_tokens`         | Post-consumption calculation   |
| `total_tokens`          | `total_tokens`              | Verification and logging       |
| `input_tokens_details`  | `prompt_tokens_details`     | Detailed token breakdown       |
| `output_tokens_details` | `completion_tokens_details` | Reasoning tokens, audio tokens |

**Input Token Estimation:** The `getResponseAPIPromptTokens()` function estimates input tokens using a 4-characters-per-token approximation for text content in the input array and instructions field.

#### Pre-consumption and Post-consumption

**Pre-consumption:** The `preConsumeResponseAPIQuota()` function reserves quota based on:

- Estimated input tokens multiplied by model ratio
- Optional max output tokens if specified in the request
- Minimum quota of 1 if calculation results in zero

**Post-consumption:** The `postConsumeResponseAPIQuota()` function calculates final billing using the same formula as ChatCompletion and follows the **DRY principle** by calling the centralized `billing.PostConsumeQuotaDetailed()` function:

```text
Final Quota = (InputTokens + OutputTokens × CompletionRatio) × ModelRatio × GroupRatio + ToolsCost
```

**Centralized Billing Integration:** Both Response API and ChatCompletion API now use the same centralized billing function `billing.PostConsumeQuotaDetailed()` which ensures:

- Consistent billing logic across all APIs
- Complete logging with all metadata fields (ElapsedTime, IsStream, SystemPromptReset)
- Proper token quota management via `model.PostConsumeTokenQuota()`
- Unified user and channel quota updates

#### Centralized Billing Architecture

**DRY Principle Implementation:** The billing system follows the DRY (Don't Repeat Yourself) principle with a centralized billing architecture:

**Core Billing Functions in `relay/billing/billing.go`:**

- `PostConsumeQuota()` - Simple billing for Audio API (legacy compatibility)
- `PostConsumeQuotaDetailed()` - Detailed billing for ChatCompletion and Response API
- `ReturnPreConsumedQuota()` - Quota refunding on errors

**Unified Billing Flow:**

1. **Pre-consumption:** Both APIs use `model.PreConsumeTokenQuota()` for quota reservation
2. **Post-consumption:** Both APIs call `billing.PostConsumeQuotaDetailed()` for final billing
3. **Logging:** Unified logging with complete metadata (ElapsedTime, IsStream, SystemPromptReset)
4. **Quota Management:** Consistent user and channel quota updates

#### Pricing Compatibility

The Response API billing uses the same **Four-Layer Pricing Resolution** system as ChatCompletion:

1. **Channel-Specific Overrides** (Highest Priority)
2. **Adapter Default Pricing** (Second Priority)
3. **Global Pricing Fallback** (Third Priority)
4. **Final Default** (Lowest Priority)

The pricing is retrieved via `pricing.GetModelRatioWithThreeLayers()` which ensures consistent pricing across both API formats.

**Billing Formula:**

The same billing formula applies with field name mapping:

```text
Final Quota = Base Token Cost + Tools Cost

Where:
Base Token Cost = (InputTokens + OutputTokens × CompletionRatio) × ModelRatio × GroupRatio
Tools Cost = Structured Output Cost + Function Call Cost + Web Search Cost
```

### Advanced Features Support

#### Streaming Implementation

**Direct Pass-through:** The `ResponseAPIDirectStreamHandler()` function in `relay/adaptor/openai/main.go` provides direct streaming support:

- Forwards Response API events directly to clients without conversion
- Accumulates response text from `response.output_text.delta` events
- Extracts usage information from `response.completed` events
- Maintains proper SSE headers and event formatting

**Event Processing:** Supports all Response API streaming event types including text deltas, reasoning summaries, and completion events

#### Structured Output Support

**Detection:** The system detects structured output requests when the Response API request includes `text.format.type: "json_schema"`

**Billing:** Applies the same 25% surcharge on completion tokens as ChatCompletion API for structured output requests

#### Function Calling Support

**Tool Format:** Supports Response API tools format for function calling

**Billing:** Function calls are billed through the standard token mechanism - tool definitions count toward input tokens, and tool responses count toward output tokens

### API Endpoints and Usage

**Main Endpoint:** `POST /v1/responses`

**Basic Request Example:**

```json
{
  "model": "gpt-4",
  "input": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ]
}
```

**Streaming Request:** Add `"stream": true` to enable streaming responses

**Channel Support:** Currently limited to OpenAI channels (channel type 1) only

### Error Handling and Validation

**Request Validation:**

- Model field validation (required)
- Input array validation (required, non-empty)
- Channel type validation (OpenAI only)

**Quota Management:**

- Pre-consumption quota validation
- Automatic quota refunding on errors via `billing.ReturnPreConsumedQuota()`
- Proper error responses in Response API format

**Upstream Errors:** Pass-through of OpenAI API errors with proper HTTP status codes

### Implementation Benefits

**Native Response API Support:**

- Users can leverage Response API features directly without conversion overhead
- Full feature parity with OpenAI's Response API
- Direct access to Response API-specific capabilities

**Billing Consistency:**

- Same billing logic and pricing system applies regardless of API format
- Accurate cost calculation for all Response API features
- Transparent pricing across all supported models

**Backward Compatibility:**

- No breaking changes to existing ChatCompletion functionality
- Existing automatic conversion remains functional for ChatCompletion requests
- Seamless integration with current infrastructure

**Performance:**

- Direct pass-through eliminates conversion overhead for Response API requests
- Efficient streaming with minimal processing
- Optimized token counting and billing

This implementation successfully provides complete Response API support while maintaining full compatibility with the existing ChatCompletion billing system and ensuring accurate cost calculation for all Response API features.
