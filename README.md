# One API

One API is an open-source API gateway that provides a unified interface for various AI models, allowing users to manage and utilize multiple AI services through a single platform. It supports token management, quota management, and usage statistics, making it easier for developers to integrate and manage different AI models in their applications.

With one-api, you can aggregate various AI APIs and request in ChatCompletion, Response, or Claude Messages API formats as needed.

![](https://s3.laisky.com/uploads/2025/07/oneapi.drawio.png)

The original author of one-api has not been active for a long time, resulting in a backlog of PRs that cannot be updated. Therefore, I forked the code and merged some PRs that I consider important. I also welcome everyone to submit PRs, and I will respond and handle them actively and quickly.

Fully compatible with the upstream version, can be used directly by replacing the container image, docker images:

- `ppcelery/one-api:latest`
- `ppcelery/one-api:arm64-latest`

Also welcome to register and use my deployed one-api gateway, which supports various mainstream models. For usage instructions, please refer to <https://wiki.laisky.com/projects/gpt/pay/cn/#page_gpt_pay_cn>.

- [One API](#one-api)
  - [Turtorial](#turtorial)
  - [New Features](#new-features)
    - [(Merged) Support gpt-vision](#merged-support-gpt-vision)
    - [Support update user's remained quota](#support-update-users-remained-quota)
    - [(Merged) Support aws claude](#merged-support-aws-claude)
    - [Support openai images edits](#support-openai-images-edits)
    - [Support gemini-2.0-flash-exp](#support-gemini-20-flash-exp)
    - [Support replicate flux \& remix](#support-replicate-flux--remix)
    - [Support replicate chat models](#support-replicate-chat-models)
    - [Support OpenAI o1/o1-mini/o1-preview](#support-openai-o1o1-minio1-preview)
    - [Get request's cost](#get-requests-cost)
    - [Support Vertex Imagen3](#support-vertex-imagen3)
    - [Support gpt-4o-audio](#support-gpt-4o-audio)
    - [Support deepseek-reasoner \& gemini-2.0-flash-thinking-exp-01-21](#support-deepseek-reasoner--gemini-20-flash-thinking-exp-01-21)
    - [Support o3-mini](#support-o3-mini)
    - [Support gemini-2.0-flash](#support-gemini-20-flash)
    - [Support OpenRouter's reasoning content](#support-openrouters-reasoning-content)
    - [Support claude-3-7-sonnet \& thinking](#support-claude-3-7-sonnet--thinking)
      - [Stream](#stream)
      - [Non-Stream](#non-stream)
    - [Automatically Enable Thinking and Customize Reasoning Format via URL Parameters](#automatically-enable-thinking-and-customize-reasoning-format-via-url-parameters)
      - [Reasoning Format - reasoning-content](#reasoning-format---reasoning-content)
      - [Reasoning Format - reasoning](#reasoning-format---reasoning)
      - [Reasoning Format - thinking](#reasoning-format---thinking)
    - [Support AWS cross-region inferences](#support-aws-cross-region-inferences)
    - [Support OpenAI web search models](#support-openai-web-search-models)
    - [Support gemini multimodal output #2197](#support-gemini-multimodal-output-2197)
    - [Support coze oauth authentication](#support-coze-oauth-authentication)
    - [Support gemini-2.5-pro](#support-gemini-25-pro)
    - [Support o3 \& o4-mini \& gpt-4.1](#support-o3--o4-mini--gpt-41)
    - [Support gpt-image-1's image generation \& edits](#support-gpt-image-1s-image-generation--edits)
    - [Support black-forest-labs/flux-kontext-pro](#support-black-forest-labsflux-kontext-pro)
    - [Support o3-pro \& reasoning content](#support-o3-pro--reasoning-content)
    - [Support GCP Vertex gloabl region and gemini-2.5-pro-preview-06-05](#support-gcp-vertex-gloabl-region-and-gemini-25-pro-preview-06-05)
    - [Support OpenAI Response API](#support-openai-response-api)
    - [Support AWS BedRock Inference Profile](#support-aws-bedrock-inference-profile)
    - [Support /v1/messages Claude Messages API](#support-v1messages-claude-messages-api)
  - [Bug fix](#bug-fix)

## Turtorial

Run one-api using docker-compose:

```yaml
oneapi:
  image: ppcelery/one-api:latest
  restart: unless-stopped
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
  environment:
    # (optional) SESSION_SECRET set a fixed session secret so that user sessions won't be invalidated after server restart
    SESSION_SECRET: xxxxxxx
    # (optional) DEBUG enable debug mode
    DEBUG: "true"
    # (optional) DEBUG_SQL display SQL logs
    DEBUG_SQL: "true"
    # (optional) If you access one-api using a non-HTTPS address, you need to set DISABLE_COOKIE_SECURE to true
    DISABLE_COOKIE_SECURE: "true"
    # (optional) ENFORCE_INCLUDE_USAGE require upstream API responses to include usage field
    ENFORCE_INCLUDE_USAGE: "true"
    # (optional) MAX_ITEMS_PER_PAGE maximum items per page, default is 10
    MAX_ITEMS_PER_PAGE: 10
    # (optional) GLOBAL_API_RATE_LIMIT maximum API requests per IP within three minutes, default is 1000
    GLOBAL_API_RATE_LIMIT: 1000
    # (optional) GLOBAL_WEB_RATE_LIMIT maximum web page requests per IP within three minutes, default is 1000
    GLOBAL_WEB_RATE_LIMIT: 1000
    # (optional) /v1 API ratelimit for each token
    GLOBAL_RELAY_RATE_LIMIT: 1000
    # (optional) Whether to ratelimit for channel, 0 is unlimited, 1 is to enable the ratelimit
    GLOBAL_CHANNEL_RATE_LIMIT: 1
    # (optional) REDIS_CONN_STRING set REDIS cache connection
    REDIS_CONN_STRING: redis://100.122.41.16:6379/1
    # (optional) FRONTEND_BASE_URL redirect page requests to specified address, server-side setting only
    FRONTEND_BASE_URL: https://oneapi.laisky.com
    # (optional) OPENROUTER_PROVIDER_SORT set sorting method for OpenRouter Providers, default is throughput
    OPENROUTER_PROVIDER_SORT: throughput
    # (optional) CHANNEL_SUSPEND_SECONDS_FOR_429 set the duration for channel suspension when receiving 429 error, default is 60 seconds
    CHANNEL_SUSPEND_SECONDS_FOR_429: 60
    # (optional) DEFAULT_MAX_TOKEN set the default maximum number of tokens for requests, default is 2048
    DEFAULT_MAX_TOKEN: 2048
  volumes:
    - /var/lib/oneapi:/data
  ports:
    - 3000:3000
```

The initial default account and password are `root` / `123456`.

## New Features

### (Merged) Support gpt-vision

### Support update user's remained quota

You can update the used quota using the API key of any token, allowing other consumption to be aggregated into the one-api for centralized management.

![](https://s3.laisky.com/uploads/2024/12/oneapi-update-quota.png)

### (Merged) Support aws claude

- [feat: support aws bedrockruntime claude3 #1328](https://github.com/songquanpeng/one-api/pull/1328)
- [feat: add new claude models #1910](https://github.com/songquanpeng/one-api/pull/1910)

![](https://s3.laisky.com/uploads/2024/12/oneapi-claude.png)

### Support openai images edits

- [feat: support openai images edits api #1369](https://github.com/songquanpeng/one-api/pull/1369)

![](https://s3.laisky.com/uploads/2024/12/oneapi-image-edit.png)

### Support gemini-2.0-flash-exp

- [feat: add gemini-2.0-flash-exp #1983](https://github.com/songquanpeng/one-api/pull/1983)

![](https://s3.laisky.com/uploads/2024/12/oneapi-gemini-flash.png)

### Support replicate flux & remix

- [feature: 支持 replicate 的绘图 #1954](https://github.com/songquanpeng/one-api/pull/1954)
- [feat: image edits/inpaiting 支持 replicate 的 flux remix #1986](https://github.com/songquanpeng/one-api/pull/1986)

![](https://s3.laisky.com/uploads/2024/12/oneapi-replicate-1.png)

![](https://s3.laisky.com/uploads/2024/12/oneapi-replicate-2.png)

![](https://s3.laisky.com/uploads/2024/12/oneapi-replicate-3.png)

### Support replicate chat models

- [feat: 支持 replicate chat models #1989](https://github.com/songquanpeng/one-api/pull/1989)

### Support OpenAI o1/o1-mini/o1-preview

- [feat: add openai o1 #1990](https://github.com/songquanpeng/one-api/pull/1990)

### Get request's cost

Each chat completion request will include a `X-Oneapi-Request-Id` in the returned headers. You can use this request id to request `GET /api/cost/request/:request_id` to get the cost of this request.

The returned structure is:

```go
type UserRequestCost struct {
  Id          int     `json:"id"`
  CreatedTime int64   `json:"created_time" gorm:"bigint"`
  UserID      int     `json:"user_id"`
  RequestID   string  `json:"request_id"`
  Quota       int64   `json:"quota"`
  CostUSD     float64 `json:"cost_usd" gorm:"-"`
}
```

### Support Vertex Imagen3

- [feat: support vertex imagen3 #2030](https://github.com/songquanpeng/one-api/pull/2030)

![](https://s3.laisky.com/uploads/2025/01/oneapi-imagen3.png)

### Support gpt-4o-audio

- [feat: support gpt-4o-audio #2032](https://github.com/songquanpeng/one-api/pull/2032)

![](https://s3.laisky.com/uploads/2025/01/oneapi-audio-1.png)

![](https://s3.laisky.com/uploads/2025/01/oneapi-audio-2.png)

### Support deepseek-reasoner & gemini-2.0-flash-thinking-exp-01-21

- [feature: add deepseek-reasoner & gemini-2.0-flash-thinking-exp-01-21 #2045](https://github.com/songquanpeng/one-api/pull/2045)

### Support o3-mini

- [feat: extend support for o3 models and update model ratios #2048](https://github.com/songquanpeng/one-api/pull/2048)

### Support gemini-2.0-flash

- [feat: support gemini-2.0-flash #2055](https://github.com/songquanpeng/one-api/pull/2055)

### Support OpenRouter's reasoning content

- [feat: support OpenRouter reasoning #2108](https://github.com/songquanpeng/one-api/pull/2108)

By default, the thinking mode is automatically enabled for the deepseek-r1 model, and the response is returned in the open-router format.

![](https://s3.laisky.com/uploads/2025/02/openrouter-reasoning.png)

### Support claude-3-7-sonnet & thinking

- [feat: support claude-3-7-sonnet #2143](https://github.com/songquanpeng/one-api/pull/2143/files)
- [feat: support claude thinking #2144](https://github.com/songquanpeng/one-api/pull/2144)

By default, the thinking mode is not enabled. You need to manually pass the `thinking` field in the request body to enable it.

#### Stream

![](https://s3.laisky.com/uploads/2025/02/claude-thinking.png)

#### Non-Stream

![](https://s3.laisky.com/uploads/2025/02/claude-thinking-non-stream.png)

### Automatically Enable Thinking and Customize Reasoning Format via URL Parameters

Supports two URL parameters: `thinking` and `reasoning_format`.

- `thinking`: Whether to enable thinking mode, disabled by default.
- `reasoning_format`: Specifies the format of the returned reasoning.
  - `reasoning_content`: DeepSeek official API format, returned in the `reasoning_content` field.
  - `reasoning`: OpenRouter format, returned in the `reasoning` field.
  - `thinking`: Claude format, returned in the `thinking` field.

#### Reasoning Format - reasoning-content

![](https://s3.laisky.com/uploads/2025/02/reasoning_format-reasoning_content.png)

#### Reasoning Format - reasoning

![](https://s3.laisky.com/uploads/2025/02/reasoning_format-reasoning.png)

#### Reasoning Format - thinking

![](https://s3.laisky.com/uploads/2025/02/reasoning_format-thinking.png)

### Support AWS cross-region inferences

- [fix: support aws cross region inferences #2182](https://github.com/songquanpeng/one-api/pull/2182)

### Support OpenAI web search models

- [feature: support openai web search models #2189](https://github.com/songquanpeng/one-api/pull/2189)

support `gpt-4o-search-preview` & `gpt-4o-mini-search-preview`

![](https://s3.laisky.com/uploads/2025/03/openai-websearch-models-1.png)

![](https://s3.laisky.com/uploads/2025/03/openai-websearch-models-2.png)

### Support gemini multimodal output #2197

- [feature: support gemini multimodal output #2197](https://github.com/songquanpeng/one-api/pull/2197)

![](https://s3.laisky.com/uploads/2025/03/gemini-multimodal.png)

### Support coze oauth authentication

- [feat: support coze oauth authentication](https://github.com/Laisky/one-api/pull/52)

### Support gemini-2.5-pro

### Support o3 & o4-mini & gpt-4.1

### Support gpt-image-1's image generation & edits

![](https://s3.laisky.com/uploads/2025/04/gpt-image-1-2.png)

![](https://s3.laisky.com/uploads/2025/04/gpt-image-1-3.png)

![](https://s3.laisky.com/uploads/2025/04/gpt-image-1-1.png)

### Support black-forest-labs/flux-kontext-pro

![](https://s3.laisky.com/uploads/2025/05/flux-kontext-pro.png)

### Support o3-pro & reasoning content

![](https://s3.laisky.com/uploads/2025/06/o3-pro.png)

### Support GCP Vertex gloabl region and gemini-2.5-pro-preview-06-05

![](https://s3.laisky.com/uploads/2025/06/gemini-2.5-pro-preview-06-05.png)

### Support OpenAI Response API

**Partially supported, still in development.**

![](https://s3.laisky.com/uploads/2025/07/response-api.png)

### Support AWS BedRock Inference Profile

![](https://s3.laisky.com/uploads/2025/07/aws-inference-profile.png)

### Support /v1/messages Claude Messages API

![](https://s3.laisky.com/uploads/2025/07/claude_messages.png)

## Bug fix

- [BUGFIX: Several issues when updating tokens #1933](https://github.com/songquanpeng/one-api/pull/1933)
- [feat(audio): count whisper-1 quota by audio duration #2022](https://github.com/songquanpeng/one-api/pull/2022)
- [fix: Fix issue where high-quota users using low-quota tokens aren't pre-charged, causing large token deficits under high concurrency #25](https://github.com/Laisky/one-api/pull/25)
- [fix: channel test false negative #2065](https://github.com/songquanpeng/one-api/pull/2065)
- [fix: resolve "bufio.Scanner: token too long" error by increasing buffer size #2128](https://github.com/songquanpeng/one-api/pull/2128)
- [feat: Enhance VolcEngine channel support with bot model #2131](https://github.com/songquanpeng/one-api/pull/2131)
- [fix: models API returns models in deactivated channels #2150](https://github.com/songquanpeng/one-api/pull/2150)
- [fix: Automatically close channel when connection fails](https://github.com/Laisky/one-api/pull/34)
- [fix: update EmailDomainWhitelist submission logic #33](https://github.com/Laisky/one-api/pull/33)
- [fix: send ByAll](https://github.com/Laisky/one-api/pull/35)
- [fix: oidc token endpoint request body #2106 #36](https://github.com/Laisky/one-api/pull/36)
