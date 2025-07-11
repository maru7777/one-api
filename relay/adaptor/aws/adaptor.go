package aws

import (
	"context"
	"io"
	"net/http"

	"github.com/Laisky/errors/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/aws/utils"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

var _ adaptor.Adaptor = new(Adaptor)

type Adaptor struct {
	awsAdapter utils.AwsAdapter
	Config     aws.Config
	Meta       *meta.Meta
	AwsClient  *bedrockruntime.Client
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.Meta = meta
	defaultConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(meta.Config.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			meta.Config.AK, meta.Config.SK, "")))
	if err != nil {
		return
	}
	a.Config = defaultConfig
	a.AwsClient = bedrockruntime.NewFromConfig(defaultConfig)
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	adaptor := GetAdaptor(request.Model)
	if adaptor == nil {
		return nil, errors.New("adaptor not found")
	}

	a.awsAdapter = adaptor
	return adaptor.ConvertRequest(c, relayMode, request)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if a.awsAdapter == nil {
		return nil, utils.WrapErr(errors.New("awsAdapter is nil"))
	}
	return a.awsAdapter.DoResponse(c, a.AwsClient, meta)
}

func (a *Adaptor) GetModelList() (models []string) {
	for model := range adaptors {
		models = append(models, model)
	}
	return
}

func (a *Adaptor) GetChannelName() string {
	return "aws"
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

// Pricing methods - AWS adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.000001

	// Direct map definition - much easier to maintain and edit
	// Pricing from https://aws.amazon.com/bedrock/pricing/
	return map[string]adaptor.ModelPrice{
		// Claude Models on AWS Bedrock
		"claude-instant-1.2":         {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 3.125}, // $0.8/$2.5 per 1K tokens
		"claude-2.0":                 {Ratio: 8 * MilliTokensUsd, CompletionRatio: 3.125},   // $8/$25 per 1K tokens
		"claude-2.1":                 {Ratio: 8 * MilliTokensUsd, CompletionRatio: 3.125},   // $8/$25 per 1K tokens
		"claude-3-haiku-20240307":    {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 5},    // $0.25/$1.25 per 1K tokens
		"claude-3-sonnet-20240229":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens
		"claude-3-opus-20240229":     {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5},      // $15/$75 per 1K tokens
		"claude-opus-4-20250514":     {Ratio: 15 * MilliTokensUsd, CompletionRatio: 5},      // $15/$75 per 1K tokens
		"claude-3-5-sonnet-20240620": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens
		"claude-3-5-sonnet-20241022": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens
		"claude-3-5-sonnet-latest":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens
		"claude-3-5-haiku-20241022":  {Ratio: 1 * MilliTokensUsd, CompletionRatio: 5},       // $1/$5 per 1K tokens
		"claude-3-7-sonnet-latest":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens
		"claude-3-7-sonnet-20250219": {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens
		"claude-sonnet-4-20250514":   {Ratio: 3 * MilliTokensUsd, CompletionRatio: 5},       // $3/$15 per 1K tokens

		// Llama Models on AWS Bedrock
		"llama3-8b-8192":  {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 2},  // $0.3/$0.6 per 1K tokens
		"llama3-70b-8192": {Ratio: 2.65 * MilliTokensUsd, CompletionRatio: 1}, // $2.65/$2.65 per 1K tokens

		// Amazon Nova Models (if supported)
		"amazon-nova-micro":   {Ratio: 0.035 * MilliTokensUsd, CompletionRatio: 4.28}, // $0.035/$0.15 per 1K tokens
		"amazon-nova-lite":    {Ratio: 0.06 * MilliTokensUsd, CompletionRatio: 4.17},  // $0.06/$0.25 per 1K tokens
		"amazon-nova-pro":     {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 4},      // $0.8/$3.2 per 1K tokens
		"amazon-nova-premier": {Ratio: 2.4 * MilliTokensUsd, CompletionRatio: 4.17},   // $2.4/$10 per 1K tokens

		// Titan Models (if supported)
		"amazon-titan-text-lite":    {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1.33}, // $0.3/$0.4 per 1K tokens
		"amazon-titan-text-express": {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 2},    // $0.8/$1.6 per 1K tokens
		"amazon-titan-embed-text":   {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},    // $0.1 per 1K tokens

		// Cohere Models (if supported)
		"cohere-command-text":       {Ratio: 1.5 * MilliTokensUsd, CompletionRatio: 1.33}, // $1.5/$2 per 1K tokens
		"cohere-command-light-text": {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 2},    // $0.3/$0.6 per 1K tokens

		// AI21 Models (if supported)
		"ai21-j2-mid":    {Ratio: 12.5 * MilliTokensUsd, CompletionRatio: 1}, // $12.5 per 1K tokens
		"ai21-j2-ultra":  {Ratio: 18.8 * MilliTokensUsd, CompletionRatio: 1}, // $18.8 per 1K tokens
		"ai21-jamba-1.5": {Ratio: 2 * MilliTokensUsd, CompletionRatio: 4},    // $2/$8 per 1K tokens

		// Mistral Models (if supported)
		"mistral-7b-instruct":   {Ratio: 0.15 * MilliTokensUsd, CompletionRatio: 1.33}, // $0.15/$0.2 per 1K tokens
		"mistral-8x7b-instruct": {Ratio: 0.45 * MilliTokensUsd, CompletionRatio: 1.56}, // $0.45/$0.7 per 1K tokens
		"mistral-large":         {Ratio: 4 * MilliTokensUsd, CompletionRatio: 3},       // $4/$12 per 1K tokens
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default AWS pricing (Claude-like)
	return 3 * 0.000001 // Default USD pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for AWS
	return 5.0
}
