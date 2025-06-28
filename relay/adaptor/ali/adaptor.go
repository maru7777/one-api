package ali

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// https://help.aliyun.com/zh/dashscope/developer-reference/api-details

type Adaptor struct {
	meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	fullRequestURL := ""
	switch meta.Mode {
	case relaymode.Embeddings:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/embeddings/text-embedding/text-embedding", meta.BaseURL)
	case relaymode.ImagesGenerations:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", meta.BaseURL)
	default:
		fullRequestURL = fmt.Sprintf("%s/api/v1/services/aigc/text-generation/generation", meta.BaseURL)
	}

	return fullRequestURL, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.IsStream {
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("X-DashScope-SSE", "enable")
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)

	if meta.Mode == relaymode.ImagesGenerations {
		req.Header.Set("X-DashScope-Async", "enable")
	}
	if a.meta.Config.Plugin != "" {
		req.Header.Set("X-DashScope-Plugin", a.meta.Config.Plugin)
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	switch relayMode {
	case relaymode.Embeddings:
		aliEmbeddingRequest := ConvertEmbeddingRequest(*request)
		return aliEmbeddingRequest, nil
	default:
		aliRequest := ConvertRequest(*request)
		return aliRequest, nil
	}
}

func (a *Adaptor) ConvertImageRequest(_ *gin.Context, request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	aliRequest := ConvertImageRequest(*request)
	return aliRequest, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		err, usage = StreamHandler(c, resp)
	} else {
		switch meta.Mode {
		case relaymode.Embeddings:
			err, usage = EmbeddingHandler(c, resp)
		case relaymode.ImagesGenerations:
			err, usage = ImageHandler(c, resp)
		default:
			err, usage = Handler(c, resp)
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "ali"
}

// Pricing methods - Ali adapter manages its own model pricing
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliRmb = 0.0001

	// Direct map definition - much easier to maintain and edit
	// Pricing from https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
	return map[string]adaptor.ModelPrice{
		// Qwen Turbo Models
		"qwen-turbo":        {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen-turbo-latest": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Qwen Plus Models
		"qwen-plus":        {Ratio: 0.8 * MilliRmb, CompletionRatio: 1},
		"qwen-plus-latest": {Ratio: 0.8 * MilliRmb, CompletionRatio: 1},

		// Qwen Max Models
		"qwen-max":             {Ratio: 2.4 * MilliRmb, CompletionRatio: 1},
		"qwen-max-latest":      {Ratio: 2.4 * MilliRmb, CompletionRatio: 1},
		"qwen-max-longcontext": {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},

		// Qwen Vision Models
		"qwen-vl-max":         {Ratio: 3 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-max-latest":  {Ratio: 3 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-plus":        {Ratio: 1.5 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-plus-latest": {Ratio: 1.5 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-ocr":         {Ratio: 5 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-ocr-latest":  {Ratio: 5 * MilliRmb, CompletionRatio: 1},

		// Qwen Audio Models
		"qwen-audio-turbo": {Ratio: 1.4286, CompletionRatio: 1},

		// Qwen Math Models
		"qwen-math-plus":         {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen-math-plus-latest":  {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen-math-turbo":        {Ratio: 2 * MilliRmb, CompletionRatio: 1},
		"qwen-math-turbo-latest": {Ratio: 2 * MilliRmb, CompletionRatio: 1},

		// Qwen Coder Models
		"qwen-coder-plus":         {Ratio: 3.5 * MilliRmb, CompletionRatio: 1},
		"qwen-coder-plus-latest":  {Ratio: 3.5 * MilliRmb, CompletionRatio: 1},
		"qwen-coder-turbo":        {Ratio: 2 * MilliRmb, CompletionRatio: 1},
		"qwen-coder-turbo-latest": {Ratio: 2 * MilliRmb, CompletionRatio: 1},

		// Qwen MT Models
		"qwen-mt-plus":  {Ratio: 15 * MilliRmb, CompletionRatio: 1},
		"qwen-mt-turbo": {Ratio: 1 * MilliRmb, CompletionRatio: 1},

		// QwQ Models
		"qwq-32b-preview": {Ratio: 2 * MilliRmb, CompletionRatio: 1},

		// Qwen 2.5 Models
		"qwen2.5-72b-instruct":  {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-32b-instruct":  {Ratio: 30 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-14b-instruct":  {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-7b-instruct":   {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-3b-instruct":   {Ratio: 6 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-1.5b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-0.5b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Qwen 2 Models
		"qwen2-72b-instruct":      {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen2-57b-a14b-instruct": {Ratio: 3.5 * MilliRmb, CompletionRatio: 1},
		"qwen2-7b-instruct":       {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen2-1.5b-instruct":     {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2-0.5b-instruct":     {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Qwen 1.5 Models
		"qwen1.5-110b-chat": {Ratio: 8 * MilliRmb, CompletionRatio: 1},
		"qwen1.5-72b-chat":  {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen1.5-32b-chat":  {Ratio: 2 * MilliRmb, CompletionRatio: 1},
		"qwen1.5-14b-chat":  {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen1.5-7b-chat":   {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen1.5-1.8b-chat": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen1.5-0.5b-chat": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Qwen 1 Models
		"qwen-72b-chat":              {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen-14b-chat":              {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen-7b-chat":               {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen-1.8b-chat":             {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen-1.8b-longcontext-chat": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// QVQ Models
		"qvq-72b-preview": {Ratio: 4 * MilliRmb, CompletionRatio: 1},

		// Qwen 2.5 VL Models
		"qwen2.5-vl-72b-instruct":  {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-vl-7b-instruct":   {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-vl-2b-instruct":   {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-vl-1b-instruct":   {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-vl-0.5b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Qwen 2 VL Models
		"qwen2-vl-7b-instruct": {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen2-vl-2b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-v1":           {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen-vl-chat-v1":      {Ratio: 1 * MilliRmb, CompletionRatio: 1},

		// Qwen Audio Models
		"qwen2-audio-instruct": {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen-audio-chat":      {Ratio: 1 * MilliRmb, CompletionRatio: 1},

		// Qwen Math Models (additional)
		"qwen2.5-math-72b-instruct":  {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-math-7b-instruct":   {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-math-1.5b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2-math-72b-instruct":    {Ratio: 4 * MilliRmb, CompletionRatio: 1},
		"qwen2-math-7b-instruct":     {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen2-math-1.5b-instruct":   {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Qwen Coder Models (additional)
		"qwen2.5-coder-32b-instruct":  {Ratio: 2 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-coder-14b-instruct":  {Ratio: 1 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-coder-7b-instruct":   {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-coder-3b-instruct":   {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-coder-1.5b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},
		"qwen2.5-coder-0.5b-instruct": {Ratio: 0.3 * MilliRmb, CompletionRatio: 1},

		// Embedding Models
		"text-embedding-v1":       {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"text-embedding-v3":       {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"text-embedding-v2":       {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"text-embedding-async-v2": {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},
		"text-embedding-async-v1": {Ratio: 0.5 * MilliRmb, CompletionRatio: 1},

		// Image Generation Models
		"ali-stable-diffusion-xl":   {Ratio: 8 * MilliRmb, CompletionRatio: 1},
		"ali-stable-diffusion-v1.5": {Ratio: 8 * MilliRmb, CompletionRatio: 1},
		"wanx-v1":                   {Ratio: 8 * MilliRmb, CompletionRatio: 1},

		// DeepSeek Models (hosted on Ali)
		"deepseek-r1":                   {Ratio: 1 * MilliRmb, CompletionRatio: 8},
		"deepseek-v3":                   {Ratio: 1 * MilliRmb, CompletionRatio: 2},
		"deepseek-r1-distill-qwen-1.5b": {Ratio: 0.07 * MilliRmb, CompletionRatio: 0.28},
		"deepseek-r1-distill-qwen-7b":   {Ratio: 0.14 * MilliRmb, CompletionRatio: 0.28},
		"deepseek-r1-distill-qwen-14b":  {Ratio: 0.28 * MilliRmb, CompletionRatio: 0.28},
		"deepseek-r1-distill-qwen-32b":  {Ratio: 0.42 * MilliRmb, CompletionRatio: 0.28},
		"deepseek-r1-distill-llama-8b":  {Ratio: 0.14 * MilliRmb, CompletionRatio: 0.28},
		"deepseek-r1-distill-llama-70b": {Ratio: 1 * MilliRmb, CompletionRatio: 2},
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default Ali pricing
	return 0.8 * 0.0001 // Default RMB pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for Ali
	return 1.0
}
