package pricing

import (
	"fmt"
	"sync"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/apitype"
)

// DefaultGlobalPricingAdapters defines which adapters contribute to global pricing fallback
// This can be easily modified to add or remove adapters from the global pricing system
// Only includes adapters that have been refactored with comprehensive pricing models
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

// GlobalPricingManager manages the third-layer global model pricing
// It merges pricing from selected adapters to provide fallback pricing
// for custom channels that don't have specific model pricing
type GlobalPricingManager struct {
	mu                   sync.RWMutex
	globalModelPricing   map[string]adaptor.ModelConfig
	contributingAdapters []int // API types of adapters to include in global pricing
	initialized          bool
	getAdaptorFunc       func(apiType int) adaptor.Adaptor
}

// Global instance of the pricing manager
var globalPricingManager = &GlobalPricingManager{
	// Will be initialized from configuration
	contributingAdapters: nil,
}

// InitializeGlobalPricingManager sets up the global pricing manager with the adaptor getter function
func InitializeGlobalPricingManager(getAdaptor func(apiType int) adaptor.Adaptor) {
	globalPricingManager.mu.Lock()
	defer globalPricingManager.mu.Unlock()

	globalPricingManager.getAdaptorFunc = getAdaptor

	// Load contributing adapters from default configuration
	if globalPricingManager.contributingAdapters == nil {
		globalPricingManager.contributingAdapters = make([]int, len(DefaultGlobalPricingAdapters))
		copy(globalPricingManager.contributingAdapters, DefaultGlobalPricingAdapters)
		logger.SysLog(fmt.Sprintf("Loaded %d adapters for global pricing",
			len(globalPricingManager.contributingAdapters)))
	}

	globalPricingManager.initialized = false // Force re-initialization with new function

	logger.SysLog("Global pricing manager initialized")
}

// SetContributingAdapters allows configuration of which adapters contribute to global pricing
func SetContributingAdapters(apiTypes []int) {
	globalPricingManager.mu.Lock()
	defer globalPricingManager.mu.Unlock()

	globalPricingManager.contributingAdapters = make([]int, len(apiTypes))
	copy(globalPricingManager.contributingAdapters, apiTypes)
	globalPricingManager.initialized = false // Force re-initialization

	logger.SysLog("Global pricing adapters updated, will reload on next access")
}

// ReloadDefaultConfiguration reloads the adapter configuration from the default slice
func ReloadDefaultConfiguration() {
	globalPricingManager.mu.Lock()
	defer globalPricingManager.mu.Unlock()

	globalPricingManager.contributingAdapters = make([]int, len(DefaultGlobalPricingAdapters))
	copy(globalPricingManager.contributingAdapters, DefaultGlobalPricingAdapters)
	globalPricingManager.initialized = false // Force re-initialization

	logger.SysLog(fmt.Sprintf("Reloaded global pricing configuration: %d adapters", len(globalPricingManager.contributingAdapters)))
}

// GetContributingAdapters returns the current list of contributing adapters
func GetContributingAdapters() []int {
	globalPricingManager.mu.RLock()
	defer globalPricingManager.mu.RUnlock()

	result := make([]int, len(globalPricingManager.contributingAdapters))
	copy(result, globalPricingManager.contributingAdapters)
	return result
}

// ensureInitialized ensures the global pricing is initialized
// Must be called with at least read lock held
func (gpm *GlobalPricingManager) ensureInitialized() {
	if gpm.initialized || gpm.getAdaptorFunc == nil {
		return
	}

	// Upgrade to write lock
	gpm.mu.RUnlock()
	gpm.mu.Lock()
	defer func() {
		gpm.mu.Unlock()
		gpm.mu.RLock()
	}()

	// Check again after acquiring write lock
	if gpm.initialized {
		return
	}

	gpm.initializeUnsafe()
}

// initializeUnsafe performs the actual initialization without locking
// Must be called with write lock held
func (gpm *GlobalPricingManager) initializeUnsafe() {
	if gpm.getAdaptorFunc == nil {
		logger.SysWarn("Global pricing manager not properly initialized - missing adaptor getter function")
		return
	}

	logger.SysLog("Initializing global model pricing from contributing adapters...")

	gpm.globalModelPricing = make(map[string]adaptor.ModelConfig)
	successCount := 0

	for _, apiType := range gpm.contributingAdapters {
		if gpm.mergeAdapterPricing(apiType) {
			successCount++
		}
	}

	gpm.initialized = true
	logger.SysLog(fmt.Sprintf("Global model pricing initialized with %d models from %d/%d adapters",
		len(gpm.globalModelPricing), successCount, len(gpm.contributingAdapters)))
}

// mergeAdapterPricing merges pricing from a specific adapter
// Must be called with write lock held
// Returns true if successful, false otherwise
func (gpm *GlobalPricingManager) mergeAdapterPricing(apiType int) bool {
	adaptor := gpm.getAdaptorFunc(apiType)
	if adaptor == nil {
		logger.SysWarn(fmt.Sprintf("No adaptor found for API type %d", apiType))
		return false
	}

	pricing := adaptor.GetDefaultModelPricing()
	if len(pricing) == 0 {
		logger.SysWarn(fmt.Sprintf("Adaptor %d returned empty pricing", apiType))
		return false
	}

	mergedCount := 0
	conflictCount := 0

	for modelName, modelPrice := range pricing {
		if existingPrice, exists := gpm.globalModelPricing[modelName]; exists {
			// Handle conflict: prefer the first adapter's pricing (could be configurable)
			logger.SysWarn(fmt.Sprintf("Model %s pricing conflict: existing=%.9f, new=%.9f (keeping existing)",
				modelName, existingPrice.Ratio, modelPrice.Ratio))
			conflictCount++
		} else {
			gpm.globalModelPricing[modelName] = modelPrice
			mergedCount++
		}
	}

	logger.SysLog(fmt.Sprintf("Merged %d models from adapter %d (%d conflicts)", mergedCount, apiType, conflictCount))
	return true
}

// GetGlobalModelRatio returns the global model ratio for a given model
// Returns 0 if the model is not found in global pricing
func GetGlobalModelRatio(modelName string) float64 {
	globalPricingManager.mu.RLock()
	defer globalPricingManager.mu.RUnlock()

	globalPricingManager.ensureInitialized()

	if price, exists := globalPricingManager.globalModelPricing[modelName]; exists {
		return price.Ratio
	}

	return 0 // Not found in global pricing
}

// GetGlobalCompletionRatio returns the global completion ratio for a given model
// Returns 0 if the model is not found in global pricing
func GetGlobalCompletionRatio(modelName string) float64 {
	globalPricingManager.mu.RLock()
	defer globalPricingManager.mu.RUnlock()

	globalPricingManager.ensureInitialized()

	if price, exists := globalPricingManager.globalModelPricing[modelName]; exists {
		return price.CompletionRatio
	}

	return 0 // Not found in global pricing
}

// GetGlobalModelPricing returns a copy of the entire global pricing map
func GetGlobalModelPricing() map[string]adaptor.ModelConfig {
	globalPricingManager.mu.RLock()
	defer globalPricingManager.mu.RUnlock()

	globalPricingManager.ensureInitialized()

	// Return a copy to prevent external modification
	result := make(map[string]adaptor.ModelConfig)
	for k, v := range globalPricingManager.globalModelPricing {
		result[k] = v
	}

	return result
}

// ReloadGlobalPricing forces a reload of the global pricing from contributing adapters
func ReloadGlobalPricing() {
	globalPricingManager.mu.Lock()
	defer globalPricingManager.mu.Unlock()

	globalPricingManager.initialized = false
	globalPricingManager.initializeUnsafe()
}

// GetGlobalPricingStats returns statistics about the global pricing
func GetGlobalPricingStats() (int, int) {
	globalPricingManager.mu.RLock()
	defer globalPricingManager.mu.RUnlock()

	globalPricingManager.ensureInitialized()

	return len(globalPricingManager.globalModelPricing), len(globalPricingManager.contributingAdapters)
}

// IsGlobalPricingInitialized returns whether the global pricing has been initialized
func IsGlobalPricingInitialized() bool {
	globalPricingManager.mu.RLock()
	defer globalPricingManager.mu.RUnlock()

	return globalPricingManager.initialized && globalPricingManager.getAdaptorFunc != nil
}

// GetModelRatioWithThreeLayers implements the three-layer pricing fallback:
// 1. Channel-specific overrides (highest priority)
// 2. Adapter default pricing (second priority)
// 3. Global pricing fallback (third priority)
// 4. Final default (lowest priority)
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

// GetCompletionRatioWithThreeLayers implements the three-layer completion ratio fallback
func GetCompletionRatioWithThreeLayers(modelName string, channelOverrides map[string]float64, adaptor adaptor.Adaptor) float64 {
	// Layer 1: User custom ratio (channel-specific overrides)
	if channelOverrides != nil {
		if override, exists := channelOverrides[modelName]; exists {
			return override
		}
	}

	// Layer 2: Channel default ratio (adapter's default pricing)
	if adaptor != nil {
		ratio := adaptor.GetCompletionRatio(modelName)
		// Check if the adapter actually has pricing for this model
		defaultPricing := adaptor.GetDefaultModelPricing()
		if _, hasSpecificPricing := defaultPricing[modelName]; hasSpecificPricing {
			return ratio
		}
	}

	// Layer 3: Global model pricing (merged from selected adapters)
	globalRatio := GetGlobalCompletionRatio(modelName)
	if globalRatio > 0 {
		return globalRatio
	}

	// Layer 4: Final fallback - reasonable default
	return 1.0 // Default completion ratio
}
