package aiproxy

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
)

// ModelRatios uses the same pricing as OpenAI since AIProxy is an OpenAI proxy
var ModelRatios = openai.ModelRatios

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
