package openrouter

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
	adaptor.DefaultPricingMethods
}

func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	return "", nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return nil, nil
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	return nil, nil
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(a.GetDefaultModelPricing())
}

func (a *Adaptor) GetChannelName() string {
	return "openrouter"
}

// GetDefaultModelPricing returns the pricing information for OpenRouter models
// Based on OpenRouter pricing: https://openrouter.ai/models
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelPrice {
	const MilliTokensUsd = 0.000001

	return map[string]adaptor.ModelPrice{
		// OpenRouter Models - Based on https://openrouter.ai/models
		"01-ai/yi-large":                                  {Ratio: 1.5 * MilliTokensUsd, CompletionRatio: 1},
		"aetherwiing/mn-starcannon-12b":                   {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"ai21/jamba-1-5-large":                            {Ratio: 4.0 * MilliTokensUsd, CompletionRatio: 1},
		"ai21/jamba-1-5-mini":                             {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"ai21/jamba-instruct":                             {Ratio: 0.35 * MilliTokensUsd, CompletionRatio: 1},
		"aion-labs/aion-1.0":                              {Ratio: 6.0 * MilliTokensUsd, CompletionRatio: 1},
		"aion-labs/aion-1.0-mini":                         {Ratio: 1.2 * MilliTokensUsd, CompletionRatio: 1},
		"aion-labs/aion-rp-llama-3.1-8b":                  {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"allenai/llama-3.1-tulu-3-405b":                   {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"alpindale/goliath-120b":                          {Ratio: 4.6875 * MilliTokensUsd, CompletionRatio: 1},
		"alpindale/magnum-72b":                            {Ratio: 1.125 * MilliTokensUsd, CompletionRatio: 1},
		"amazon/nova-lite-v1":                             {Ratio: 0.12 * MilliTokensUsd, CompletionRatio: 1},
		"amazon/nova-micro-v1":                            {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"amazon/nova-pro-v1":                              {Ratio: 1.6 * MilliTokensUsd, CompletionRatio: 1},
		"anthracite-org/magnum-v2-72b":                    {Ratio: 1.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthracite-org/magnum-v4-72b":                    {Ratio: 1.125 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-2":                              {Ratio: 12.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-2.0":                            {Ratio: 12.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-2.0:beta":                       {Ratio: 12.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-2.1":                            {Ratio: 12.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-2.1:beta":                       {Ratio: 12.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-2:beta":                         {Ratio: 12.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3-haiku":                        {Ratio: 0.625 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3-haiku:beta":                   {Ratio: 0.625 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.5-haiku-20241022":             {Ratio: 2.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.5-haiku-20241022:beta":        {Ratio: 2.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.5-haiku:beta":                 {Ratio: 2.0 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3-sonnet":                       {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3-sonnet:beta":                  {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.5-sonnet:beta":                {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.5-sonnet-20240620":            {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.5-sonnet-20240620:beta":       {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3.7-sonnet:beta":                {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3-opus":                         {Ratio: 37.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-3-opus:beta":                    {Ratio: 37.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-4-opus":                         {Ratio: 37.5 * MilliTokensUsd, CompletionRatio: 1},
		"anthropic/claude-4-opus:beta":                    {Ratio: 37.5 * MilliTokensUsd, CompletionRatio: 1},
		"cognitivecomputations/dolphin-mixtral-8x22b":     {Ratio: 0.45 * MilliTokensUsd, CompletionRatio: 1},
		"cognitivecomputations/dolphin-mixtral-8x7b":      {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command":                                  {Ratio: 0.95 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r":                                {Ratio: 0.7125 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r-03-2024":                        {Ratio: 0.7125 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r-08-2024":                        {Ratio: 0.285 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r-plus":                           {Ratio: 7.125 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r-plus-04-2024":                   {Ratio: 7.125 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r-plus-08-2024":                   {Ratio: 4.75 * MilliTokensUsd, CompletionRatio: 1},
		"cohere/command-r7b-12-2024":                      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 1},
		"databricks/dbrx-instruct":                        {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-chat":                          {Ratio: 1.25 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-chat-v2.5":                     {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-chat:free":                     {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1":                            {Ratio: 7 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1-distill-llama-70b":          {Ratio: 0.345 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1-distill-llama-70b:free":     {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1-distill-llama-8b":           {Ratio: 0.02 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1-distill-qwen-1.5b":          {Ratio: 0.09 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1-distill-qwen-14b":           {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1-distill-qwen-32b":           {Ratio: 0.09 * MilliTokensUsd, CompletionRatio: 1},
		"deepseek/deepseek-r1:free":                       {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"eva-unit-01/eva-llama-3.33-70b":                  {Ratio: 3.0 * MilliTokensUsd, CompletionRatio: 1},
		"eva-unit-01/eva-qwen-2.5-32b":                    {Ratio: 1.7 * MilliTokensUsd, CompletionRatio: 1},
		"eva-unit-01/eva-qwen-2.5-72b":                    {Ratio: 3.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-2.0-flash-001":                     {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-2.0-flash-exp:free":                {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-2.0-flash-lite-preview-02-05:free": {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-2.0-flash-thinking-exp-1219:free":  {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-2.0-flash-thinking-exp:free":       {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-2.0-pro-exp-02-05:free":            {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-exp-1206:free":                     {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-flash-1.5":                         {Ratio: 0.15 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-flash-1.5-8b":                      {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-flash-1.5-8b-exp":                  {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-pro":                               {Ratio: 0.75 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-pro-1.5":                           {Ratio: 2.5 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemini-pro-vision":                        {Ratio: 0.75 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-27b-it":                           {Ratio: 0.135 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-9b-it":                            {Ratio: 0.03 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-2-9b-it:free":                       {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/gemma-7b-it":                              {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 1},
		"google/learnlm-1.5-pro-experimental:free":        {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/palm-2-chat-bison":                        {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/palm-2-chat-bison-32k":                    {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/palm-2-codechat-bison":                    {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"google/palm-2-codechat-bison-32k":                {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"gryphe/mythomax-l2-13b":                          {Ratio: 0.0325 * MilliTokensUsd, CompletionRatio: 1},
		"gryphe/mythomax-l2-13b:free":                     {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"huggingfaceh4/zephyr-7b-beta:free":               {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"infermatic/mn-inferor-12b":                       {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"inflection/inflection-3-pi":                      {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"inflection/inflection-3-productivity":            {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"jondurbin/airoboros-l2-70b":                      {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 1},
		"liquid/lfm-3b":                                   {Ratio: 0.01 * MilliTokensUsd, CompletionRatio: 1},
		"liquid/lfm-40b":                                  {Ratio: 0.075 * MilliTokensUsd, CompletionRatio: 1},
		"liquid/lfm-7b":                                   {Ratio: 0.005 * MilliTokensUsd, CompletionRatio: 1},
		"mancer/weaver":                                   {Ratio: 1.125 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-2-13b-chat":                     {Ratio: 0.11 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-2-70b-chat":                     {Ratio: 0.45 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3-70b-instruct":                 {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3-8b-instruct":                  {Ratio: 0.03 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3-8b-instruct:free":             {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.1-405b":                       {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.1-405b-instruct":              {Ratio: 0.4 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.1-70b-instruct":               {Ratio: 0.15 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.1-8b-instruct":                {Ratio: 0.025 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.2-11b-vision-instruct":        {Ratio: 0.0275 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.2-11b-vision-instruct:free":   {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.2-1b-instruct":                {Ratio: 0.005 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.2-3b-instruct":                {Ratio: 0.0125 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.2-90b-vision-instruct":        {Ratio: 0.8 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.3-70b-instruct":               {Ratio: 0.15 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-3.3-70b-instruct:free":          {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"meta-llama/llama-guard-2-8b":                     {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/phi-3-medium-128k-instruct":            {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/phi-3-medium-128k-instruct:free":       {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/phi-3-mini-128k-instruct":              {Ratio: 0.05 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/phi-3-mini-128k-instruct:free":         {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/phi-3.5-mini-128k-instruct":            {Ratio: 0.05 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/phi-4":                                 {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/wizardlm-2-7b":                         {Ratio: 0.035 * MilliTokensUsd, CompletionRatio: 1},
		"microsoft/wizardlm-2-8x22b":                      {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 1},
		"minimax/minimax-01":                              {Ratio: 0.55 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/codestral-2501":                        {Ratio: 0.45 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/codestral-mamba":                       {Ratio: 0.125 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/ministral-3b":                          {Ratio: 0.02 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/ministral-8b":                          {Ratio: 0.05 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-7b-instruct":                   {Ratio: 0.0275 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-7b-instruct-v0.1":              {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-7b-instruct-v0.3":              {Ratio: 0.0275 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-7b-instruct:free":              {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-large":                         {Ratio: 3.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-large-2407":                    {Ratio: 3.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-large-2411":                    {Ratio: 3.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-medium":                        {Ratio: 4.05 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-nemo":                          {Ratio: 0.04 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-nemo:free":                     {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-small":                         {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-small-24b-instruct-2501":       {Ratio: 0.07 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-small-24b-instruct-2501:free":  {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mistral-tiny":                          {Ratio: 0.125 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mixtral-8x22b-instruct":                {Ratio: 0.45 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mixtral-8x7b":                          {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/mixtral-8x7b-instruct":                 {Ratio: 0.12 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/pixtral-12b":                           {Ratio: 0.05 * MilliTokensUsd, CompletionRatio: 1},
		"mistralai/pixtral-large-2411":                    {Ratio: 3.0 * MilliTokensUsd, CompletionRatio: 1},
		"neversleep/llama-3-lumimaid-70b":                 {Ratio: 2.25 * MilliTokensUsd, CompletionRatio: 1},
		"neversleep/llama-3-lumimaid-8b":                  {Ratio: 0.5625 * MilliTokensUsd, CompletionRatio: 1},
		"neversleep/llama-3-lumimaid-8b:extended":         {Ratio: 0.5625 * MilliTokensUsd, CompletionRatio: 1},
		"neversleep/llama-3.1-lumimaid-70b":               {Ratio: 2.25 * MilliTokensUsd, CompletionRatio: 1},
		"neversleep/llama-3.1-lumimaid-8b":                {Ratio: 0.5625 * MilliTokensUsd, CompletionRatio: 1},
		"neversleep/noromaid-20b":                         {Ratio: 1.125 * MilliTokensUsd, CompletionRatio: 1},
		"nothingiisreal/mn-celeste-12b":                   {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"nousresearch/hermes-2-pro-llama-3-8b":            {Ratio: 0.02 * MilliTokensUsd, CompletionRatio: 1},
		"nousresearch/hermes-3-llama-3.1-405b":            {Ratio: 0.4 * MilliTokensUsd, CompletionRatio: 1},
		"nousresearch/hermes-3-llama-3.1-70b":             {Ratio: 0.15 * MilliTokensUsd, CompletionRatio: 1},
		"nousresearch/nous-hermes-2-mixtral-8x7b-dpo":     {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1},
		"nousresearch/nous-hermes-llama2-13b":             {Ratio: 0.085 * MilliTokensUsd, CompletionRatio: 1},
		"nvidia/llama-3.1-nemotron-70b-instruct":          {Ratio: 0.15 * MilliTokensUsd, CompletionRatio: 1},
		"nvidia/llama-3.1-nemotron-70b-instruct:free":     {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/chatgpt-4o-latest":                        {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-3.5-turbo":                            {Ratio: 0.75 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-3.5-turbo-0125":                       {Ratio: 0.75 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-3.5-turbo-0613":                       {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-3.5-turbo-1106":                       {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-3.5-turbo-16k":                        {Ratio: 2.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-3.5-turbo-instruct":                   {Ratio: 1.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4":                                    {Ratio: 30.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4-0314":                               {Ratio: 30.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4-1106-preview":                       {Ratio: 15.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4-32k":                                {Ratio: 60.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4-32k-0314":                           {Ratio: 60.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4-turbo":                              {Ratio: 15.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4-turbo-preview":                      {Ratio: 15.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o":                                   {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o-2024-05-13":                        {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o-2024-08-06":                        {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o-2024-11-20":                        {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o-mini":                              {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o-mini-2024-07-18":                   {Ratio: 0.3 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4o:extended":                          {Ratio: 9.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/gpt-4.5-preview":                          {Ratio: 75 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o1":                                       {Ratio: 30.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o1-mini":                                  {Ratio: 2.2 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o1-mini-2024-09-12":                       {Ratio: 2.2 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o1-preview":                               {Ratio: 30.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o1-preview-2024-09-12":                    {Ratio: 30.0 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o3-mini":                                  {Ratio: 2.2 * MilliTokensUsd, CompletionRatio: 1},
		"openai/o3-mini-high":                             {Ratio: 2.2 * MilliTokensUsd, CompletionRatio: 1},
		"openchat/openchat-7b":                            {Ratio: 0.0275 * MilliTokensUsd, CompletionRatio: 1},
		"openchat/openchat-7b:free":                       {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/llama-3.1-sonar-huge-128k-online":     {Ratio: 2.5 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/llama-3.1-sonar-large-128k-chat":      {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/llama-3.1-sonar-large-128k-online":    {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/llama-3.1-sonar-small-128k-chat":      {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/llama-3.1-sonar-small-128k-online":    {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/sonar":                                {Ratio: 0.5 * MilliTokensUsd, CompletionRatio: 1},
		"perplexity/sonar-reasoning":                      {Ratio: 2.5 * MilliTokensUsd, CompletionRatio: 1},
		"pygmalionai/mythalion-13b":                       {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qvq-72b-preview":                            {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-72b-instruct":                        {Ratio: 0.45 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-7b-instruct":                         {Ratio: 0.027 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-7b-instruct:free":                    {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-vl-72b-instruct":                     {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2-vl-7b-instruct":                      {Ratio: 0.05 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2.5-72b-instruct":                      {Ratio: 0.2 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2.5-7b-instruct":                       {Ratio: 0.025 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-2.5-coder-32b-instruct":                {Ratio: 0.08 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-max":                                   {Ratio: 3.2 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-plus":                                  {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-turbo":                                 {Ratio: 0.1 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen-vl-plus:free":                          {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwen2.5-vl-72b-instruct:free":               {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"qwen/qwq-32b-preview":                            {Ratio: 0.09 * MilliTokensUsd, CompletionRatio: 1},
		"raifle/sorcererlm-8x22b":                         {Ratio: 2.25 * MilliTokensUsd, CompletionRatio: 1},
		"sao10k/fimbulvetr-11b-v2":                        {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"sao10k/l3-euryale-70b":                           {Ratio: 0.4 * MilliTokensUsd, CompletionRatio: 1},
		"sao10k/l3-lunaris-8b":                            {Ratio: 0.03 * MilliTokensUsd, CompletionRatio: 1},
		"sao10k/l3.1-70b-hanami-x1":                       {Ratio: 1.5 * MilliTokensUsd, CompletionRatio: 1},
		"sao10k/l3.1-euryale-70b":                         {Ratio: 0.4 * MilliTokensUsd, CompletionRatio: 1},
		"sao10k/l3.3-euryale-70b":                         {Ratio: 0.4 * MilliTokensUsd, CompletionRatio: 1},
		"sophosympatheia/midnight-rose-70b":               {Ratio: 0.4 * MilliTokensUsd, CompletionRatio: 1},
		"sophosympatheia/rogue-rose-103b-v0.2:free":       {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"teknium/openhermes-2.5-mistral-7b":               {Ratio: 0.085 * MilliTokensUsd, CompletionRatio: 1},
		"thedrummer/rocinante-12b":                        {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 1},
		"thedrummer/unslopnemo-12b":                       {Ratio: 0.25 * MilliTokensUsd, CompletionRatio: 1},
		"undi95/remm-slerp-l2-13b":                        {Ratio: 0.6 * MilliTokensUsd, CompletionRatio: 1},
		"undi95/toppy-m-7b":                               {Ratio: 0.035 * MilliTokensUsd, CompletionRatio: 1},
		"undi95/toppy-m-7b:free":                          {Ratio: 0.0 * MilliTokensUsd, CompletionRatio: 1},
		"x-ai/grok-2-1212":                                {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"x-ai/grok-2-vision-1212":                         {Ratio: 5.0 * MilliTokensUsd, CompletionRatio: 1},
		"x-ai/grok-beta":                                  {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"x-ai/grok-vision-beta":                           {Ratio: 7.5 * MilliTokensUsd, CompletionRatio: 1},
		"xwin-lm/xwin-lm-70b":                             {Ratio: 1.875 * MilliTokensUsd, CompletionRatio: 1},
	}
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetModelRatio(modelName)
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetCompletionRatio(modelName)
}
