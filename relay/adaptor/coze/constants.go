package coze

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Coze models are typically free or very low cost for basic usage
var ModelRatios = map[string]adaptor.ModelPrice{
	// Coze models - estimated pricing as Coze doesn't publish detailed pricing
	"coze-chat": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

const (
	PersonalAccessToken = "personal_access_token"
	OAuthJWT            = "oauth_jwt"
)
