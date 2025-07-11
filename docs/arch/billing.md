# One-API Billing System Architecture

## Table of Contents

- [One-API Billing System Architecture](#one-api-billing-system-architecture)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
    - [Key Features](#key-features)
  - [System Architecture](#system-architecture)
  - [Core Components](#core-components)
    - [1. Quota Management System](#1-quota-management-system)
      - [Key Files](#key-files)
    - [2. Pricing System](#2-pricing-system)
      - [Key Files](#key-files-1)
    - [3. Adapter System](#3-adapter-system)
      - [Adapter Pricing Implementation Status](#adapter-pricing-implementation-status)
  - [Quota Management](#quota-management)
    - [User Quota System](#user-quota-system)
      - [Database Schema (User)](#database-schema-user)
      - [Key Operations](#key-operations)
    - [Token Quota System](#token-quota-system)
      - [Database Schema (Token)](#database-schema-token)
      - [Token Types](#token-types)
    - [Caching Strategy](#caching-strategy)
      - [Key Files](#key-files-2)
  - [Pricing System](#pricing-system)
    - [Pricing Hierarchy](#pricing-hierarchy)
    - [Pricing Constants](#pricing-constants)
    - [Model Pricing Structure](#model-pricing-structure)
    - [Global Pricing System](#global-pricing-system)
      - [Architecture](#architecture)
      - [Configuration](#configuration)
      - [Key Functions](#key-functions)
      - [Conflict Resolution](#conflict-resolution)
    - [Channel-specific Pricing](#channel-specific-pricing)
      - [Database Schema (Channel)](#database-schema-channel)
      - [Key Functions](#key-functions-1)
  - [Billing Flow](#billing-flow)
    - [Request Processing Flow](#request-processing-flow)
    - [Pre-consumption Phase](#pre-consumption-phase)
      - [Key Functions](#key-functions-2)
    - [Post-consumption Phase](#post-consumption-phase)
      - [Key Functions](#key-functions-3)
    - [Quota Calculation](#quota-calculation)
      - [Text Requests](#text-requests)
      - [Audio Requests](#audio-requests)
      - [Image Requests](#image-requests)
  - [Database Schema](#database-schema)
    - [Core Tables](#core-tables)
      - [Users Table](#users-table)
      - [Tokens Table](#tokens-table)
      - [Channels Table](#channels-table)
      - [Logs Table](#logs-table)
    - [Relationships](#relationships)
  - [API Endpoints](#api-endpoints)
    - [Billing Endpoints](#billing-endpoints)
      - [Get Subscription Info](#get-subscription-info)
      - [Get Usage Info](#get-usage-info)
    - [Channel Pricing Endpoints](#channel-pricing-endpoints)
      - [Get Channel Pricing](#get-channel-pricing)
      - [Update Channel Pricing](#update-channel-pricing)
      - [Get Default Pricing](#get-default-pricing)
    - [Token Management Endpoints](#token-management-endpoints)
  - [Recent Pricing Implementation Improvements](#recent-pricing-implementation-improvements)
    - [Comprehensive Adapter Pricing Implementation](#comprehensive-adapter-pricing-implementation)
      - [Implementation Overview](#implementation-overview)
      - [Key Improvements](#key-improvements)
      - [Technical Implementation Details](#technical-implementation-details)
      - [Provider-Specific Pricing Examples](#provider-specific-pricing-examples)
      - [Migration from DefaultPricingMethods](#migration-from-defaultpricingmethods)
      - [Impact and Benefits](#impact-and-benefits)
      - [Files Modified](#files-modified)
    - [Global Pricing Enhancement and Clean Architecture](#global-pricing-enhancement-and-clean-architecture)
      - [Architecture Transformation](#architecture-transformation)
      - [What Was Enhanced](#what-was-enhanced)
      - [What Was Preserved](#what-was-preserved)
      - [Benefits Achieved](#benefits-achieved)
  - [Implementation Details](#implementation-details)
    - [Four-Layer Pricing Resolution](#four-layer-pricing-resolution)
      - [Legacy Compatibility](#legacy-compatibility)
    - [Batch Processing](#batch-processing)
      - [Configuration](#configuration-1)
      - [Batch Types](#batch-types)
      - [Key Files](#key-files-3)
    - [Error Handling and Recovery](#error-handling-and-recovery)
      - [Quota Refund Mechanism](#quota-refund-mechanism)
      - [Key Files](#key-files-4)
    - [Structured Output Pricing](#structured-output-pricing)
      - [Key Files](#key-files-5)
  - [Performance Considerations](#performance-considerations)
    - [Caching Strategy](#caching-strategy-1)
    - [Database Optimization](#database-optimization)
    - [Memory Management](#memory-management)
    - [Monitoring and Metrics](#monitoring-and-metrics)
      - [Key Metrics](#key-metrics)
      - [Monitoring Files](#monitoring-files)
  - [Summary](#summary)

## Overview

The One-API billing system is a comprehensive quota and pricing management system designed to handle multi-tenant API usage billing across various AI model providers. The system supports both user-level and token-level quota management, channel-specific pricing, and real-time billing calculations.

**Current System Status (2025)**: The billing system now features comprehensive pricing coverage with 25+ adapters supporting 400+ models across major AI providers, implementing a sophisticated four-layer pricing resolution system with intelligent global fallback capabilities.

### Key Features

- **Multi-tier Quota System**: User quotas, token quotas, and unlimited quota support
- **Channel-specific Pricing**: Per-channel model pricing overrides
- **Comprehensive Adapter Pricing**: 25+ channel adapters with native pricing implementations
- **Complete Pricing Coverage**: 400+ models across major AI providers with accurate pricing
- **Four-Layer Pricing System**: Channel overrides → Adapter pricing → Global fallback → Final default
- **Real-time Billing**: Pre-consumption and post-consumption quota management
- **Caching Layer**: Redis-based caching for performance optimization
- **Batch Updates**: Configurable batch processing for high-throughput scenarios
- **Zero-Loss Pricing Display**: All completion ratios (including 0) are displayed in the UI

## System Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        API[API Requests]
        WEB[Web Dashboard]
    end

    subgraph "Controller Layer"
        TC[Text Controller]
        AC[Audio Controller]
        IC[Image Controller]
        BC[Billing Controller]
        CC[Channel Controller]
    end

    subgraph "Billing Core"
        QM[Quota Manager]
        PM[Pricing Manager]
        BF[Billing Flow]
    end

    subgraph "Adapter Layer"
        OA[OpenAI Adapter]
        AA[Anthropic Adapter]
        GA[Gemini Adapter]
        ZA[Zhipu Adapter]
        ALA[Ali Adapter]
        BA[Baidu Adapter]
        TA[Tencent Adapter]
        XA[Xunfei Adapter]
        VA[VertexAI Adapter]
        AWA[AWS Adapter]
        RA[Replicate Adapter]
        CHA[Cohere Adapter]
        CFA[Cloudflare Adapter]
        CA[Other Adapters]
    end

    subgraph "Data Layer"
        DB[(Database)]
        CACHE[(Redis Cache)]
        BATCH[Batch Processor]
    end

    API --> TC
    API --> AC
    API --> IC
    WEB --> BC
    WEB --> CC

    TC --> QM
    AC --> QM
    IC --> QM

    QM --> PM
    PM --> OA
    PM --> AA
    PM --> GA
    PM --> ZA
    PM --> ALA
    PM --> BA
    PM --> TA
    PM --> XA
    PM --> VA
    PM --> AWA
    PM --> RA
    PM --> CHA
    PM --> CFA
    PM --> CA

    QM --> DB
    QM --> CACHE
    QM --> BATCH

    BF --> DB
    BF --> CACHE
```

## Core Components

### 1. Quota Management System

The quota management system operates on multiple levels:

```mermaid
graph LR
    subgraph "Quota Hierarchy"
        U[User Quota] --> T[Token Quota]
        T --> R[Request Quota]
    end

    subgraph "Quota Types"
        RQ[Remaining Quota]
        UQ[Used Quota]
        UNL[Unlimited Quota]
    end

    subgraph "Operations"
        PRE[Pre-consume]
        POST[Post-consume]
        REFUND[Refund]
    end
```

#### Key Files

- `model/user.go` - User quota management
- `model/token.go` - Token quota management
- `model/cache.go` - Quota caching layer
- `relay/billing/billing.go` - Core billing operations

### 2. Pricing System

The pricing system uses a hierarchical approach with three fallback levels:

```mermaid
graph TD
    REQ[Request] --> CCP{Channel-specific<br/>Override?}
    CCP -->|Yes| CSP[Layer 1: Use Channel<br/>Override]
    CCP -->|No| ADP{Adapter has<br/>Model Pricing?}

    ADP -->|Yes| ADP_PRICE[Layer 2: Use Adapter<br/>Default Pricing]
    ADP -->|No| GP{Global Pricing<br/>Available?}

    GP -->|Yes| GP_PRICE[Layer 3: Use Global<br/>Pricing Fallback]
    GP -->|No| DEFAULT[Layer 4: Use Final<br/>Default]

    CSP --> CALC[Calculate Cost]
    ADP_PRICE --> CALC
    GP_PRICE --> CALC
    DEFAULT --> CALC

    CALC --> BILL[Billing]

    style REQ fill:#e1f5fe
    style CSP fill:#c8e6c9
    style ADP_PRICE fill:#c8e6c9
    style GP_PRICE fill:#fff9c4
    style DEFAULT fill:#ffcdd2
    style CALC fill:#fff3e0
    style BILL fill:#fce4ec
```

#### Key Files

- `relay/pricing/global.go` - Global pricing manager and four-layer pricing logic
- `relay/billing/ratio/model.go` - Audio/video pricing constants and legacy compatibility functions
- `relay/adaptor/interface.go` - Adapter pricing interface
- `relay/adaptor/*/constants.go` - Adapter-specific pricing implementations (25+ adapters)
- `model/channel.go` - Channel-specific pricing storage
- `controller/channel.go` - Channel pricing API endpoints

### 3. Adapter System

Each channel adapter implements its own comprehensive pricing logic. As of the latest implementation, 25+ major adapters have native pricing support:

```mermaid
classDiagram
    class Adaptor {
        <<interface>>
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    class DefaultPricingMethods {
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    class OpenAIAdaptor {
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    class AnthropicAdaptor {
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    class AliAdaptor {
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    class GeminiAdaptor {
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    class ReplicateAdaptor {
        +GetDefaultModelPricing() map[string]ModelPrice
        +GetModelRatio(modelName string) float64
        +GetCompletionRatio(modelName string) float64
    }

    Adaptor <|-- DefaultPricingMethods
    Adaptor <|-- OpenAIAdaptor
    Adaptor <|-- AnthropicAdaptor
    Adaptor <|-- AliAdaptor
    Adaptor <|-- GeminiAdaptor
    Adaptor <|-- ReplicateAdaptor

    note for Adaptor "25+ adapters with native pricing:\n• OpenAI (84 models)\n• Anthropic (15 models)\n• Zhipu (23 models)\n• Ali (89 models)\n• Baidu (16 models)\n• Tencent (6 models)\n• Gemini (26 models)\n• Xunfei (6 models)\n• VertexAI (34 models)\n• DeepSeek (2 models)\n• Groq (20+ models)\n• Mistral (10+ models)\n• Moonshot (3 models)\n• Cohere (12 models)\n• AI360 (4 models)\n• Doubao (20+ models)\n• Novita (40+ models)\n• OpenRouter (100+ models)\n• Replicate (48 models)\n• AWS (31 models)\n• StepFun (3 models)\n• LingYi WanWu (3 models)\n• Minimax (3 models)\n• Baichuan (2 models)\n• TogetherAI (40+ models)\n• SiliconFlow (30+ models)"
```

#### Adapter Pricing Implementation Status

**✅ Adapters with Native Pricing (25+ total)**:

- **OpenAI**: 84 models with comprehensive GPT pricing
- **Anthropic**: 15 models with Claude pricing
- **Zhipu**: 23 models with GLM pricing
- **Ali (Alibaba)**: 89 models with Qwen and other models
- **Baidu**: 16 models with ERNIE pricing
- **Tencent**: 6 models with Hunyuan pricing
- **Gemini**: 26 models with Gemini/Gemma pricing
- **Xunfei**: 6 models with Spark pricing
- **VertexAI**: 34 models with Google Cloud pricing
- **DeepSeek**: 2 models with DeepSeek pricing
- **Groq**: 20+ models with Groq pricing
- **Mistral**: 10+ models with Mistral pricing
- **Moonshot**: 3 models with Moonshot pricing
- **Cohere**: 12 models with Command pricing
- **AI360**: 4 models with AI360 pricing
- **Doubao**: 20+ models with Doubao pricing
- **Novita**: 40+ models with Novita pricing
- **OpenRouter**: 100+ models with comprehensive multi-provider pricing
- **Replicate**: 48 models with image generation and language models
- **AWS**: 31 models with Bedrock pricing
- **StepFun**: 3 models with Step pricing
- **LingYi WanWu**: 3 models with Yi pricing
- **Minimax**: 3 models with abab pricing
- **Baichuan**: 2 models with Baichuan pricing
- **TogetherAI**: 40+ models with Together AI pricing
- **SiliconFlow**: 30+ models with SiliconFlow pricing
- **XAI**: 2 models with Grok pricing

**❌ Adapters Using DefaultPricingMethods (3 remaining)**:

- **Ollama**: Local model hosting (typically free)
- **Coze**: Conversational AI platform
- **DeepL**: Translation service

## Quota Management

### User Quota System

Users have a primary quota that serves as the main billing account:

#### Database Schema (User)

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    quota BIGINT DEFAULT 0,           -- Remaining quota
    used_quota BIGINT DEFAULT 0,     -- Total used quota
    request_count INTEGER DEFAULT 0   -- Total request count
);
```

#### Key Operations

- `IncreaseUserQuota()` - Add quota to user account
- `DecreaseUserQuota()` - Deduct quota from user account
- `GetUserQuota()` - Retrieve current user quota
- `CacheGetUserQuota()` - Cached quota retrieval

### Token Quota System

Tokens can have individual quotas or inherit from user quotas:

#### Database Schema (Token)

```sql
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY,
    user_id INTEGER,
    remain_quota BIGINT DEFAULT 0,    -- Token-specific quota
    used_quota BIGINT DEFAULT 0,      -- Token usage
    unlimited_quota BOOLEAN DEFAULT FALSE  -- Bypass quota checks
);
```

#### Token Types

1. **Limited Tokens**: Have specific quota limits
2. **Unlimited Tokens**: Bypass quota restrictions
3. **User-inherited Tokens**: Use parent user's quota

### Caching Strategy

The system implements a multi-layer caching strategy for performance:

```mermaid
graph LR
    REQ[Request] --> CACHE{Cache Hit?}
    CACHE -->|Yes| RETURN[Return Cached]
    CACHE -->|No| DB[Query Database]
    DB --> UPDATE[Update Cache]
    UPDATE --> RETURN

    subgraph "Cache Types"
        UC[User Quota Cache]
        TC[Token Cache]
        GC[Group Cache]
    end
```

#### Key Files

- `model/cache.go` - Caching implementations
- Cache TTL configurations in `model/cache.go`

## Pricing System

### Pricing Hierarchy

The system uses a **comprehensive four-layer pricing hierarchy** to handle custom channels and unknown models:

1. **User Custom Ratio** (Channel-specific overrides) - Highest Priority
2. **Channel Default Ratio** (Adapter's default pricing) - Second Priority
3. **Global Pricing Fallback** (Merged from selected adapters) - Third Priority
4. **Final Default** (Reasonable fallback) - Lowest Priority

This four-layer approach ensures that custom channels with common models can automatically receive appropriate pricing even when the channel adapter doesn't have specific pricing for those models. The global pricing system merges pricing from 13 major adapters to provide comprehensive fallback coverage.

### Pricing Constants

```go
// Currency and token conversion constants
const (
    QuotaPerUsd     = 500000 // $1 = 500,000 quota
    MilliTokensUsd  = 0.5    // 0.5 quota per milli-token (0.000001 USD * 500000)
    ImageUsdPerPic  = 1000   // 1000 quota per image (0.002 USD * 500000)
    MilliTokensRmb  = 3.5    // 3.5 quota per milli-token (0.000007 RMB * 500000)
    ImageRmbPerPic  = 7000   // 7000 quota per image (0.014 RMB * 500000)
    MilliTokensYuan = 3.5    // 3.5 quota per milli-token (0.000007 Yuan * 500000)
    ImageYuanPerPic = 7000   // 7000 quota per image (0.014 Yuan * 500000)
    TokensPerSec    = 10     // Video tokens per second for video generation models
)
```

### Model Pricing Structure

```go
type ModelPrice struct {
    Ratio           float64 `json:"ratio"`
    CompletionRatio float64 `json:"completion_ratio,omitempty"`
}
```

### Global Pricing System

The global pricing system provides a third layer of pricing fallback for custom channels that offer common models but don't have specific pricing defined in their adapters.

#### Architecture

The global pricing manager automatically merges pricing from selected adapters on startup:

```go
type GlobalPricingManager struct {
    globalModelPricing    map[string]adaptor.ModelPrice
    contributingAdapters  []int // API types to include
    getAdaptorFunc        func(apiType int) adaptor.Adaptor
}
```

#### Configuration

Global pricing adapters are defined as a simple slice in the code for easy modification:

```go
// DefaultGlobalPricingAdapters defines which adapters contribute to global pricing fallback
var DefaultGlobalPricingAdapters = []int{
    apitype.OpenAI,    // Comprehensive GPT models with pricing
    apitype.Anthropic, // Claude models with pricing
    apitype.Gemini,    // Google Gemini models with pricing
    apitype.Ali,       // Alibaba Qwen models with pricing
    apitype.Baidu,     // Baidu ERNIE models with pricing
    apitype.Zhipu,     // Zhipu GLM models with pricing
    apitype.DeepSeek,  // DeepSeek models with pricing
    apitype.Groq,      // Groq models with pricing
    apitype.Mistral,   // Mistral models with pricing
    apitype.Moonshot,  // Moonshot models with pricing
    apitype.Cohere,    // Cohere models with pricing
    apitype.Tencent,   // Tencent Hunyuan models with pricing
    apitype.Xunfei,    // Xunfei Spark models with pricing
}
```

#### Key Functions

- `GetModelRatioWithThreeLayers()` - Four-layer pricing resolution
- `GetCompletionRatioWithThreeLayers()` - Four-layer completion ratio resolution
- `SetContributingAdapters()` - Configure which adapters contribute to global pricing
- `ReloadGlobalPricing()` - Force reload of global pricing from adapters
- `GetGlobalPricingStats()` - Get statistics about global pricing coverage

#### Conflict Resolution

When multiple adapters define pricing for the same model, the first adapter in the configuration list takes precedence. Conflicts are logged for transparency.

### Channel-specific Pricing

Channels can override default pricing for specific models:

#### Database Schema (Channel)

```sql
CREATE TABLE channels (
    id INTEGER PRIMARY KEY,
    model_ratio TEXT,        -- JSON: {"model": ratio}
    completion_ratio TEXT    -- JSON: {"model": completion_ratio}
);
```

#### Key Functions

- `GetModelRatio()` - Retrieve channel model pricing
- `SetModelRatio()` - Update channel model pricing
- `GetCompletionRatio()` - Retrieve completion pricing
- `SetCompletionRatio()` - Update completion pricing

## Billing Flow

### Request Processing Flow

```mermaid
sequenceDiagram
    participant Client
    participant Controller
    participant QuotaManager
    participant PricingManager
    participant Adapter
    participant Database
    participant Cache

    Client->>Controller: API Request
    Controller->>PricingManager: Get Model Pricing
    PricingManager->>Adapter: Get Default Pricing
    Adapter-->>PricingManager: Return Pricing
    PricingManager-->>Controller: Final Pricing

    Controller->>QuotaManager: Pre-consume Quota
    QuotaManager->>Cache: Check User Quota
    Cache-->>QuotaManager: Quota Info
    QuotaManager->>Database: Reserve Quota
    Database-->>QuotaManager: Confirmation
    QuotaManager-->>Controller: Pre-consumption OK

    Controller->>Adapter: Process Request
    Adapter-->>Controller: Response + Usage

    Controller->>QuotaManager: Post-consume Quota
    QuotaManager->>Database: Final Billing
    QuotaManager->>Cache: Update Cache
    Database-->>QuotaManager: Confirmation
    QuotaManager-->>Controller: Billing Complete

    Controller-->>Client: API Response
```

### Pre-consumption Phase

Before processing requests, the system reserves quota:

#### Key Functions

- `preConsumeQuota()` in `relay/controller/text.go`
- `getPreConsumedQuota()` for quota calculation
- Token validation and quota checking

### Post-consumption Phase

After request completion, final billing is calculated:

#### Key Functions

- `postConsumeQuota()` in `relay/controller/helper.go`
- `PostConsumeQuota()` in `relay/billing/billing.go`
- Usage logging and metrics recording

### Quota Calculation

Different request types use different calculation methods:

#### Text Requests

```
quota = (prompt_tokens + completion_tokens * completion_ratio) * model_ratio * group_ratio
```

#### Audio Requests

```
quota = audio_duration_seconds * audio_tokens_per_second * model_ratio * group_ratio
```

#### Image Requests

```
quota = image_count * image_cost_per_pic * model_ratio * group_ratio
```

## Database Schema

### Core Tables

#### Users Table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(30) UNIQUE,
    quota BIGINT DEFAULT 0,
    used_quota BIGINT DEFAULT 0,
    request_count INTEGER DEFAULT 0,
    group VARCHAR(32) DEFAULT 'default'
);
```

#### Tokens Table

```sql
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    key CHAR(48) UNIQUE,
    remain_quota BIGINT DEFAULT 0,
    used_quota BIGINT DEFAULT 0,
    unlimited_quota BOOLEAN DEFAULT FALSE,
    status INTEGER DEFAULT 1,
    expired_time BIGINT DEFAULT -1
);
```

#### Channels Table

```sql
CREATE TABLE channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type INTEGER DEFAULT 0,
    name VARCHAR(255),
    used_quota BIGINT DEFAULT 0,
    balance FLOAT DEFAULT 0,
    model_ratio TEXT,           -- JSON string
    completion_ratio TEXT,      -- JSON string
    priority BIGINT DEFAULT 0
);
```

#### Logs Table

```sql
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    channel_id INTEGER,
    model_name VARCHAR(255),
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    quota INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Relationships

```mermaid
erDiagram
    USERS ||--o{ TOKENS : owns
    USERS ||--o{ LOGS : generates
    CHANNELS ||--o{ LOGS : processes
    TOKENS ||--o{ LOGS : authenticates

    USERS {
        int id PK
        string username
        bigint quota
        bigint used_quota
        int request_count
        string group
    }

    TOKENS {
        int id PK
        int user_id FK
        string key
        bigint remain_quota
        bigint used_quota
        boolean unlimited_quota
        int status
    }

    CHANNELS {
        int id PK
        int type
        string name
        bigint used_quota
        float balance
        text model_ratio
        text completion_ratio
        bigint priority
    }

    LOGS {
        int id PK
        int user_id FK
        int channel_id FK
        string model_name
        int prompt_tokens
        int completion_tokens
        int quota
        timestamp created_at
    }
```

## API Endpoints

### Billing Endpoints

#### Get Subscription Info

```
GET /dashboard/billing/subscription
GET /v1/dashboard/billing/subscription
```

**Controller**: `controller.GetSubscription()`
**File**: `controller/billing.go`

#### Get Usage Info

```
GET /dashboard/billing/usage
GET /v1/dashboard/billing/usage
```

**Controller**: `controller.GetUsage()`
**File**: `controller/billing.go`

### Channel Pricing Endpoints

#### Get Channel Pricing

```
GET /api/channel/pricing/:id
```

**Controller**: `controller.GetChannelPricing()`
**File**: `controller/channel.go`

#### Update Channel Pricing

```
PUT /api/channel/pricing/:id
```

**Controller**: `controller.UpdateChannelPricing()`
**File**: `controller/channel.go`

#### Get Default Pricing

```
GET /api/channel/default-pricing?type=:channelType
```

**Controller**: `controller.GetChannelDefaultPricing()`
**File**: `controller/channel.go`

**Response Format**:

```json
{
  "success": true,
  "message": "",
  "data": {
    "model_ratio": "{\"model1\": 0.001, \"model2\": 0.002}",
    "completion_ratio": "{\"model1\": 1.0, \"model2\": 3.0}"
  }
}
```

**Key Implementation Details**:

- Converts channel type to API type using `channeltype.ToAPIType()`
- Includes ALL completion ratios (including 0) for complete pricing transparency
- Returns comprehensive pricing for all models supported by the adapter

### Token Management Endpoints

Token management is handled through existing user and admin endpoints with quota operations integrated.

## Recent Pricing Implementation Improvements

### Comprehensive Adapter Pricing Implementation

As of the latest update, the system has undergone a major overhaul to implement comprehensive pricing across all major channel adapters. This addresses the previous issue where only OpenAI channels displayed pricing information.

#### Implementation Overview

```mermaid
graph TD
    subgraph "Before: Limited Pricing"
        B1[OpenAI Only] --> B2[3 Adapters with Pricing]
        B2 --> B3[~70 Models Covered]
    end

    subgraph "After: Comprehensive Pricing"
        A1[13 Major Adapters] --> A2[Native Pricing Implementation]
        A2 --> A3[350+ Models Covered]
        A3 --> A4[Complete UI Display]
    end

    B3 -.->|Major Upgrade| A1
```

#### Key Improvements

1. **Expanded Adapter Coverage**: From 3 to 25+ adapters with native pricing
2. **Model Coverage**: From ~70 to 400+ models with accurate pricing
3. **UI Consistency**: All channel edit pages now display pricing information
4. **Pricing Accuracy**: Based on official provider documentation
5. **Complete Data Display**: All completion ratios (including 0) are shown
6. **Four-Layer Pricing System**: Comprehensive fallback system with global pricing support

#### Technical Implementation Details

**Adapter Pricing Structure**:

```go
// Each adapter implements comprehensive pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
    const MilliTokensUsd = 0.000001

    return map[string]adaptor.ModelPrice{
        "model-name": {
            Ratio:           0.001 * MilliTokensUsd,  // Input pricing
            CompletionRatio: 3.0,                     // Output multiplier
        },
        // ... comprehensive model coverage
    }
}
```

**Channel Type Mapping Fix**:

```go
// Fixed the channel type to API type conversion
apiType := channeltype.ToAPIType(channelType)
adaptor := relay.GetAdaptor(apiType)
```

**Complete Pricing Display**:

```go
// Include ALL completion ratios (including 0)
for model, price := range defaultPricing {
    modelRatios[model] = price.Ratio
    completionRatios[model] = price.CompletionRatio  // No filtering
}
```

#### Provider-Specific Pricing Examples

**Ali (Alibaba Cloud)**: 89 models

- Qwen models: ¥0.0003-¥0.0024 per 1K tokens
- DeepSeek models: ¥0.0001-¥0.008 per 1K tokens
- Embedding models: ¥0.00005 per 1K tokens

**AWS Bedrock**: 31 models

- Claude models: $0.25-$75 per 1M tokens
- Llama models: $0.3-$2.65 per 1M tokens
- Amazon Nova models: $0.035-$10 per 1M tokens

**Replicate**: 48 models

- FLUX image generation: $0.003-$0.12 per image
- Language models: $0.05-$9.5 per 1M tokens
- Specialized models for various use cases

#### Migration from DefaultPricingMethods

**Before**:

```go
type Adaptor struct {
    adaptor.DefaultPricingMethods  // Empty pricing
}
```

**After**:

```go
type Adaptor struct {
    // No DefaultPricingMethods embedding
}

func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
    // Comprehensive pricing implementation
    return map[string]adaptor.ModelPrice{
        // ... detailed pricing for all supported models
    }
}
```

#### Impact and Benefits

1. **Complete Pricing Transparency**: Users can now see accurate default pricing for all major providers
2. **Improved User Experience**: Channel edit pages display comprehensive pricing information
3. **Accurate Billing**: Pricing based on official provider documentation ensures accurate cost calculations
4. **Easier Channel Management**: Default pricing serves as a reference for custom pricing configuration
5. **Reduced Support Burden**: Clear pricing information reduces user confusion and support requests

#### Files Modified

**Core Implementation Files**:

- `controller/channel.go` - Fixed channel type mapping and completion ratio filtering
- `relay/adaptor/*/adaptor.go` - Added comprehensive pricing to 13 major adapters

**Adapters with Native Pricing**:

- `relay/adaptor/openai/constants.go` - 84 OpenAI GPT models
- `relay/adaptor/anthropic/constants.go` - 15 Anthropic Claude models
- `relay/adaptor/ali/constants.go` - 89 Alibaba Cloud models
- `relay/adaptor/baidu/constants.go` - 16 Baidu ERNIE models
- `relay/adaptor/tencent/constants.go` - 6 Tencent Hunyuan models
- `relay/adaptor/geminiOpenaiCompatible/constants.go` - 26 Google Gemini models
- `relay/adaptor/xunfei/constants.go` - 6 iFlytek Spark models
- `relay/adaptor/vertexai/adaptor.go` - 34 Google Cloud VertexAI models
- `relay/adaptor/zhipu/constants.go` - 23 Zhipu GLM models
- `relay/adaptor/deepseek/constants.go` - 2 DeepSeek models
- `relay/adaptor/groq/constants.go` - 20+ Groq models
- `relay/adaptor/mistral/constants.go` - 10+ Mistral models
- `relay/adaptor/moonshot/constants.go` - 3 Moonshot models
- `relay/adaptor/cohere/constant.go` - 12 Cohere Command models
- `relay/adaptor/ai360/constants.go` - 4 AI360 models
- `relay/adaptor/doubao/constants.go` - 20+ Doubao models
- `relay/adaptor/novita/constants.go` - 40+ Novita models
- `relay/adaptor/openrouter/adaptor.go` - 100+ OpenRouter models
- `relay/adaptor/replicate/constant.go` - 48 Replicate models
- `relay/adaptor/aws/adaptor.go` - 31 AWS Bedrock models
- `relay/adaptor/stepfun/constants.go` - 3 StepFun models
- `relay/adaptor/lingyiwanwu/constants.go` - 3 LingYi WanWu models
- `relay/adaptor/minimax/constants.go` - 3 Minimax models
- `relay/adaptor/baichuan/constants.go` - 2 Baichuan models
- `relay/adaptor/togetherai/constants.go` - 40+ TogetherAI models
- `relay/adaptor/siliconflow/constants.go` - 30+ SiliconFlow models
- `relay/adaptor/xai/constants.go` - 2 XAI Grok models

### Global Pricing Enhancement and Clean Architecture

Following the comprehensive adapter pricing implementation, the system underwent a second major improvement: **enhanced global pricing system** with a clean four-layer architecture that provides comprehensive fallback coverage.

#### Architecture Transformation

```mermaid
graph TD
    subgraph "Before: Limited Coverage"
        C1[Channel Override] --> C2[Few Adapters with Pricing]
        C2 --> C3[Large Global ModelRatio]
        C3 --> C4[Fallback Default]
    end

    subgraph "After: Comprehensive Four-Layer"
        A1[Channel Override] --> A2[16+ Adapters with Pricing]
        A2 --> A3[Smart Global Fallback]
        A3 --> A4[Reasonable Default]
    end

    C4 -.->|Enhanced| A1

    style A1 fill:#c8e6c9
    style A2 fill:#c8e6c9
    style A3 fill:#fff9c4
    style A4 fill:#fff3e0
```

#### What Was Enhanced

1. **Adapter Coverage** - Expanded from 3 to 25+ adapters with native pricing
2. **Global Pricing System** - Smart merging from selected adapters instead of static maps
3. **Model Coverage** - Increased from ~70 to 400+ models with accurate pricing
4. **Pricing Resolution** - Four-layer system with comprehensive fallback coverage

#### What Was Preserved

1. **Audio/Video pricing** - Special pricing for audio and video models
2. **Currency constants** - Essential conversion constants
3. **Legacy compatibility** - Functions preserved for backward compatibility
4. **All existing functionality** - No breaking changes to existing features

#### Benefits Achieved

1. **Comprehensive Coverage**: 400+ models across 25+ adapters with accurate pricing
2. **Smart Fallback**: Global pricing system merges from selected adapters automatically
3. **Maintainability**: Each adapter manages its own pricing independently
4. **Flexibility**: Four-layer system handles any pricing scenario
5. **Type safety**: Structured pricing with `ModelPrice` interface
6. **Custom Channel Support**: Automatic pricing for custom channels with common models
7. **Configurable Fallback**: Easily configurable global pricing adapters
8. **Conflict Resolution**: Transparent handling of pricing conflicts between adapters
9. **Performance**: Efficient pricing lookups with caching and smart resolution

## Implementation Details

### Current Implementation Status

**✅ Fully Implemented Features**:

- **Four-Layer Pricing System**: Complete implementation with channel overrides → adapter pricing → global fallback → final default
- **25+ Adapter Pricing**: All major adapters have native pricing in their `constants.go` files
- **Global Pricing Manager**: Automatic merging from 13 selected adapters for comprehensive fallback coverage
- **Legacy Compatibility**: Backward compatibility maintained through legacy functions in `relay/billing/ratio/model.go`
- **Real-time Usage**: Active usage in `relay/controller/text.go`, `relay/controller/audio.go`, and `controller/channel-test.go`

**📍 Current Usage Pattern**:

```go
// Modern approach used throughout the codebase
pricingAdaptor := relay.GetAdaptor(meta.ChannelType)
modelRatio := pricing.GetModelRatioWithThreeLayers(textRequest.Model, channelModelRatio, pricingAdaptor)
completionRatio := pricing.GetCompletionRatioWithThreeLayers(textRequest.Model, channelCompletionRatio, pricingAdaptor)
```

### Four-Layer Pricing Resolution

The current pricing resolution follows a comprehensive four-layer approach:

```go
// Modern approach: Use the four-layer pricing system
func getModelPricing(modelName string, channelType int, channelOverrides map[string]float64) float64 {
    apiType := channeltype.ToAPIType(channelType)
    adaptor := relay.GetAdaptor(apiType)

    // Use the four-layer pricing system
    return pricing.GetModelRatioWithThreeLayers(modelName, channelOverrides, adaptor)
}

// Four-layer implementation:
func GetModelRatioWithThreeLayers(modelName string, channelOverrides map[string]float64, adaptor adaptor.Adaptor) float64 {
    // Layer 1: User custom ratio (channel-specific overrides)
    if channelOverrides != nil {
        if override, exists := channelOverrides[modelName]; exists {
            return override
        }
    }

    // Layer 2: Channel default ratio (adapter's default pricing)
    if adaptor != nil {
        ratio := adaptor.GetModelRatio(modelName)
        // Check if the adapter actually has pricing for this model
        // If GetModelRatio returns the default fallback, we should try global pricing
        defaultPricing := adaptor.GetDefaultModelPricing()
        if _, hasSpecificPricing := defaultPricing[modelName]; hasSpecificPricing {
            return ratio
        }
    }

    // Layer 3: Global model pricing (merged from selected adapters)
    globalRatio := GetGlobalModelRatio(modelName)
    if globalRatio > 0 {
        return globalRatio
    }

    // Layer 4: Final fallback - reasonable default
    return 2.5 * 0.000001 // 2.5 USD per million tokens
}
```

#### Legacy Compatibility

The legacy `GetModelRatioWithChannel()` function is preserved for backward compatibility but simplified:

```go
// LEGACY FUNCTION - simplified for compatibility only
func GetModelRatioWithChannel(modelName string, channelType int, channelRatio map[string]float64) float64 {
    // Check channel-specific pricing if provided
    if channelRatio != nil {
        if ratio, exists := channelRatio[modelName]; exists {
            return ratio
        }
    }

    // Legacy fallback: only audio models and reasonable default
    // Note: Global pricing maps have been removed
    return 2.5 * MilliTokensUsd
}
```

### Batch Processing

For high-throughput scenarios, the system supports batch updates:

#### Configuration

```go
config.BatchUpdateEnabled = true
```

#### Batch Types

- `BatchUpdateTypeUserQuota` - User quota updates
- `BatchUpdateTypeTokenQuota` - Token quota updates
- `BatchUpdateTypeUsedQuota` - Usage tracking
- `BatchUpdateTypeRequestCount` - Request counting

#### Key Files

- `model/batch.go` - Batch processing implementation
- Batch operations in `model/user.go` and `model/token.go`

### Error Handling and Recovery

#### Quota Refund Mechanism

When requests fail, pre-consumed quota is refunded:

```go
func ReturnPreConsumedQuota(ctx context.Context, quota int64, tokenId int) {
    if quota > 0 {
        err := PostConsumeTokenQuota(tokenId, -quota)
        if err != nil {
            logger.Error(ctx, "Failed to return pre-consumed quota: " + err.Error())
        }
    }
}
```

#### Key Files

- `relay/billing/billing.go` - Quota refund operations
- Error handling in controller files

### Structured Output Pricing

Special handling for structured output requests with additional cost multipliers:

```go
// Apply 25% additional cost for structured output
structuredOutputCost := int64(math.Ceil(float64(completionTokens) * 0.25 * modelRatio))
usage.ToolsCost += structuredOutputCost
```

#### Key Files

- `relay/adaptor/openai/adaptor.go` - Structured output cost calculation
- Test files: `relay/adaptor/openai/structured_output_*_test.go`

## Performance Considerations

### Caching Strategy

1. **User Quota Caching**: 5-minute TTL for user quota information
2. **Token Caching**: 10-minute TTL for token validation
3. **Group Model Caching**: 30-minute TTL for group permissions

### Database Optimization

1. **Indexes**: Strategic indexing on frequently queried fields
2. **Batch Updates**: Configurable batch processing for write operations
3. **Connection Pooling**: Optimized database connection management

### Memory Management

1. **Pricing Cache**: In-memory caching of pricing information
2. **Model Lists**: Cached model availability per channel
3. **Configuration Cache**: Cached system configuration

### Monitoring and Metrics

The system includes comprehensive monitoring:

#### Key Metrics

- Request rate and response time
- Quota consumption patterns
- Channel utilization
- Error rates and types

#### Monitoring Files

- `common/metrics/` - Metrics collection
- Monitoring integration in controller files

## Summary

The One-API billing system has evolved into a sophisticated, multi-layered pricing and quota management system that provides:

**✅ Complete Pricing Coverage**: 25+ adapters with 400+ models across all major AI providers
**✅ Intelligent Fallback**: Four-layer pricing resolution ensures no model is left without pricing
**✅ Flexible Configuration**: Channel-specific overrides, adapter defaults, and global fallbacks
**✅ High Performance**: Caching, batch processing, and optimized database operations
**✅ Legacy Compatibility**: Smooth migration path from old centralized pricing system
**✅ Real-time Billing**: Pre-consumption and post-consumption quota management with refund capabilities

The system successfully migrated from a centralized model ratio approach to a distributed adapter-based pricing system while maintaining full backward compatibility and adding comprehensive global pricing fallback capabilities.
