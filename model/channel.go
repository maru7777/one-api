package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
)

const (
	ChannelStatusUnknown          = 0
	ChannelStatusEnabled          = 1 // don't use 0, 0 is the default value!
	ChannelStatusManuallyDisabled = 2 // also don't use 0
	ChannelStatusAutoDisabled     = 3
)

type Channel struct {
	Id                 int     `json:"id"`
	Type               int     `json:"type" gorm:"default:0"`
	Key                string  `json:"key" gorm:"type:text"`
	Status             int     `json:"status" gorm:"default:1"`
	Name               string  `json:"name" gorm:"index"`
	Weight             *uint   `json:"weight" gorm:"default:0"`
	CreatedTime        int64   `json:"created_time" gorm:"bigint"`
	TestTime           int64   `json:"test_time" gorm:"bigint"`
	ResponseTime       int     `json:"response_time"` // in milliseconds
	BaseURL            *string `json:"base_url" gorm:"column:base_url;default:''"`
	Other              *string `json:"other"`   // DEPRECATED: please save config to field Config
	Balance            float64 `json:"balance"` // in USD
	BalanceUpdatedTime int64   `json:"balance_updated_time" gorm:"bigint"`
	Models             string  `json:"models"`
	ModelConfigs       *string `json:"model_configs" gorm:"type:text;default:''"`
	Group              string  `json:"group" gorm:"type:varchar(32);default:'default'"`
	UsedQuota          int64   `json:"used_quota" gorm:"bigint;default:0"`
	ModelMapping       *string `json:"model_mapping" gorm:"type:text;default:''"`
	Priority           *int64  `json:"priority" gorm:"bigint;default:0"`
	Config             string  `json:"config"`
	SystemPrompt       *string `json:"system_prompt" gorm:"type:text"`
	RateLimit          *int    `json:"ratelimit" gorm:"column:ratelimit;default:0"`
	// Channel-specific pricing tables
	// DEPRECATED: Use ModelConfigs instead. These fields are kept for backward compatibility and migration.
	ModelRatio      *string `json:"model_ratio" gorm:"type:text"`      // DEPRECATED: JSON string of model pricing ratios
	CompletionRatio *string `json:"completion_ratio" gorm:"type:text"` // DEPRECATED: JSON string of completion pricing ratios
	// AWS-specific configuration
	InferenceProfileArnMap *string `json:"inference_profile_arn_map" gorm:"type:text"` // JSON string mapping model names to AWS Bedrock Inference Profile ARNs
}

type ChannelConfig struct {
	Region            string `json:"region,omitempty"`
	SK                string `json:"sk,omitempty"`
	AK                string `json:"ak,omitempty"`
	UserID            string `json:"user_id,omitempty"`
	APIVersion        string `json:"api_version,omitempty"`
	LibraryID         string `json:"library_id,omitempty"`
	Plugin            string `json:"plugin,omitempty"`
	VertexAIProjectID string `json:"vertex_ai_project_id,omitempty"`
	VertexAIADC       string `json:"vertex_ai_adc,omitempty"`
	AuthType          string `json:"auth_type,omitempty"`
}

type ModelConfig struct {
	MaxTokens int32 `json:"max_tokens,omitempty"`
}

// ModelConfigLocal represents the local definition of ModelConfig to avoid import cycles
// This should match the structure in relay/adaptor/interface.go
type ModelConfigLocal struct {
	Ratio           float64 `json:"ratio"`
	CompletionRatio float64 `json:"completion_ratio,omitempty"`
	MaxTokens       int32   `json:"max_tokens,omitempty"`
}

func GetAllChannels(startIdx int, num int, scope string) ([]*Channel, error) {
	var channels []*Channel
	var err error
	switch scope {
	case "all":
		err = DB.Order("id desc").Find(&channels).Error
	case "disabled":
		err = DB.Order("id desc").Where("status = ? or status = ?", ChannelStatusAutoDisabled, ChannelStatusManuallyDisabled).Find(&channels).Error
	default:
		err = DB.Order("id desc").Limit(num).Offset(startIdx).Omit("key").Find(&channels).Error
	}
	return channels, err
}

func SearchChannels(keyword string) (channels []*Channel, err error) {
	err = DB.Omit("key").Where("id = ? or name LIKE ?", helper.String2Int(keyword), keyword+"%").Find(&channels).Error
	return channels, err
}

func GetChannelById(id int, selectAll bool) (*Channel, error) {
	channel := Channel{Id: id}
	var err error = nil
	if selectAll {
		err = DB.First(&channel, "id = ?", id).Error
	} else {
		err = DB.Omit("key").First(&channel, "id = ?", id).Error
	}
	return &channel, err
}

func BatchInsertChannels(channels []Channel) error {
	var err error
	err = DB.Create(&channels).Error
	if err != nil {
		return err
	}
	for _, channel_ := range channels {
		err = channel_.AddAbilities()
		if err != nil {
			return err
		}
	}
	InitChannelCache()
	return nil
}

func (channel *Channel) GetPriority() int64 {
	if channel.Priority == nil {
		return 0
	}
	return *channel.Priority
}

func (channel *Channel) GetBaseURL() string {
	if channel.BaseURL == nil {
		return ""
	}
	return *channel.BaseURL
}

func (channel *Channel) GetModelMapping() map[string]string {
	if channel.ModelMapping == nil || *channel.ModelMapping == "" || *channel.ModelMapping == "{}" {
		return nil
	}
	modelMapping := make(map[string]string)
	err := json.Unmarshal([]byte(*channel.ModelMapping), &modelMapping)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal model mapping for channel %d, error: %s", channel.Id, err.Error()))
		return nil
	}
	return modelMapping
}

// GetModelConfig returns the model configuration for a specific model
// DEPRECATED: Use GetModelPriceConfig() instead. This method is kept for backward compatibility.
func (channel *Channel) GetModelConfig(modelName string) *ModelConfig {
	// Only use unified ModelConfigs after migration
	priceConfig := channel.GetModelPriceConfig(modelName)
	if priceConfig != nil {
		// Convert ModelPriceLocal to ModelConfig for backward compatibility
		return &ModelConfig{
			MaxTokens: priceConfig.MaxTokens,
		}
	}

	return nil
}

// MigrateModelConfigsToModelPrice migrates existing ModelConfigs data from the old format
// (map[string]ModelConfig) to the new format (map[string]ModelPriceLocal)
// This handles cases where contributors have already applied the PR changes locally
func (channel *Channel) MigrateModelConfigsToModelPrice() error {
	if channel.ModelConfigs == nil || *channel.ModelConfigs == "" || *channel.ModelConfigs == "{}" {
		return nil // Nothing to migrate
	}

	// Validate JSON format first
	var rawData interface{}
	if err := json.Unmarshal([]byte(*channel.ModelConfigs), &rawData); err != nil {
		logger.SysError(fmt.Sprintf("Channel %d has invalid JSON in ModelConfigs: %s", channel.Id, err.Error()))
		return fmt.Errorf("invalid JSON in ModelConfigs: %w", err)
	}

	// Check if the JSON is null, array, or string (invalid types)
	switch rawData.(type) {
	case nil:
		logger.SysError(fmt.Sprintf("Channel %d ModelConfigs cannot be parsed: null value", channel.Id))
		return fmt.Errorf("ModelConfigs cannot be parsed: null value")
	case []interface{}:
		logger.SysError(fmt.Sprintf("Channel %d ModelConfigs cannot be parsed: array value", channel.Id))
		return fmt.Errorf("ModelConfigs cannot be parsed: array value")
	case string:
		logger.SysError(fmt.Sprintf("Channel %d ModelConfigs cannot be parsed: string value", channel.Id))
		return fmt.Errorf("ModelConfigs cannot be parsed: string value")
	}

	// Try to unmarshal as the new format first
	var newFormatConfigs map[string]ModelConfigLocal
	err := json.Unmarshal([]byte(*channel.ModelConfigs), &newFormatConfigs)
	if err == nil {
		// Validate the new format data
		if err := channel.validateModelPriceConfigs(newFormatConfigs); err != nil {
			logger.SysError(fmt.Sprintf("Channel %d has invalid ModelPriceLocal data: %s", channel.Id, err.Error()))
			return fmt.Errorf("invalid ModelPriceLocal data: %w", err)
		}

		// Check if it has pricing data (already in new format)
		hasPricingData := false
		for _, config := range newFormatConfigs {
			if config.Ratio != 0 || config.CompletionRatio != 0 {
				hasPricingData = true
				break
			}
		}

		if hasPricingData {
			logger.SysLog(fmt.Sprintf("Channel %d ModelConfigs already in new format with pricing data", channel.Id))
			return nil
		}

		logger.SysLog(fmt.Sprintf("Channel %d ModelConfigs in new format but needs pricing migration", channel.Id))
	}

	// Try to unmarshal as the old format (map[string]ModelConfig)
	var oldFormatConfigs map[string]ModelConfig
	err = json.Unmarshal([]byte(*channel.ModelConfigs), &oldFormatConfigs)
	if err != nil {
		logger.SysError(fmt.Sprintf("Channel %d ModelConfigs cannot be parsed in either format: %s", channel.Id, err.Error()))
		return fmt.Errorf("ModelConfigs cannot be parsed in either format: %w", err)
	}

	// Validate old format data
	for modelName, config := range oldFormatConfigs {
		if modelName == "" {
			logger.SysError(fmt.Sprintf("Channel %d has empty model name in ModelConfigs", channel.Id))
			return fmt.Errorf("empty model name found in ModelConfigs")
		}
		if config.MaxTokens < 0 {
			logger.SysError(fmt.Sprintf("Channel %d has negative MaxTokens for model %s", channel.Id, modelName))
			return fmt.Errorf("negative MaxTokens for model %s", modelName)
		}
	}

	// Convert old format to new format
	migratedConfigs := make(map[string]ModelConfigLocal)

	// Get existing ModelRatio and CompletionRatio for this channel
	modelRatios := channel.GetModelRatio()
	completionRatios := channel.GetCompletionRatio()

	// Collect all model names from all sources
	allModelNames := make(map[string]bool)
	for modelName := range oldFormatConfigs {
		if modelName != "" {
			allModelNames[modelName] = true
		}
	}
	if modelRatios != nil {
		for modelName := range modelRatios {
			if modelName != "" {
				allModelNames[modelName] = true
			}
		}
	}
	if completionRatios != nil {
		for modelName := range completionRatios {
			if modelName != "" {
				allModelNames[modelName] = true
			}
		}
	}

	// Process all models from all sources
	for modelName := range allModelNames {
		newConfig := ModelConfigLocal{}

		// Start with MaxTokens from old config if available
		if oldConfig, exists := oldFormatConfigs[modelName]; exists {
			newConfig.MaxTokens = oldConfig.MaxTokens
		}

		// Add pricing information if available
		if modelRatios != nil {
			if ratio, exists := modelRatios[modelName]; exists {
				if ratio < 0 {
					return fmt.Errorf("negative ratio for model %s: %f", modelName, ratio)
				}
				if ratio > 0 {
					newConfig.Ratio = ratio
				}
			}
		}
		if completionRatios != nil {
			if completionRatio, exists := completionRatios[modelName]; exists {
				if completionRatio < 0 {
					return fmt.Errorf("negative completion ratio for model %s: %f", modelName, completionRatio)
				}
				if completionRatio > 0 {
					newConfig.CompletionRatio = completionRatio
				}
			}
		}

		migratedConfigs[modelName] = newConfig
	}

	// Validate migrated data
	if err := channel.validateModelPriceConfigs(migratedConfigs); err != nil {
		logger.SysError(fmt.Sprintf("Channel %d migration produced invalid data: %s", channel.Id, err.Error()))
		return fmt.Errorf("migration produced invalid data: %w", err)
	}

	// Save the migrated data back to ModelConfigs
	jsonBytes, err := json.Marshal(migratedConfigs)
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to marshal migrated ModelConfigs for channel %d: %s", channel.Id, err.Error()))
		return fmt.Errorf("failed to marshal migrated data: %w", err)
	}

	jsonStr := string(jsonBytes)
	channel.ModelConfigs = &jsonStr

	logger.SysLog(fmt.Sprintf("Successfully migrated ModelConfigs for channel %d from old format to new format (%d models)", channel.Id, len(migratedConfigs)))
	return nil
}

// validateModelPriceConfigs validates the structure and values of ModelPriceLocal configurations
func (channel *Channel) validateModelPriceConfigs(configs map[string]ModelConfigLocal) error {
	if configs == nil {
		return nil
	}

	for modelName, config := range configs {
		// Validate model name
		if modelName == "" {
			return fmt.Errorf("empty model name found")
		}

		// Validate ratio values
		if config.Ratio < 0 {
			return fmt.Errorf("negative ratio for model %s: %f", modelName, config.Ratio)
		}
		if config.CompletionRatio < 0 {
			return fmt.Errorf("negative completion ratio for model %s: %f", modelName, config.CompletionRatio)
		}

		// Validate MaxTokens
		if config.MaxTokens < 0 {
			return fmt.Errorf("negative MaxTokens for model %s: %d", modelName, config.MaxTokens)
		}

		// Validate that at least one field has meaningful data
		if config.Ratio == 0 && config.CompletionRatio == 0 && config.MaxTokens == 0 {
			return fmt.Errorf("model %s has no meaningful configuration data", modelName)
		}
	}

	return nil
}

// GetModelPriceConfigs returns the channel-specific model price configurations in the new unified format
func (channel *Channel) GetModelPriceConfigs() map[string]ModelConfigLocal {
	if channel.ModelConfigs == nil || *channel.ModelConfigs == "" || *channel.ModelConfigs == "{}" {
		return nil
	}

	modelPriceConfigs := make(map[string]ModelConfigLocal)
	err := json.Unmarshal([]byte(*channel.ModelConfigs), &modelPriceConfigs)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal model price configs for channel %d, error: %s", channel.Id, err.Error()))
		return nil
	}

	return modelPriceConfigs
}

// SetModelPriceConfigs sets the channel-specific model price configurations in the new unified format
func (channel *Channel) SetModelPriceConfigs(modelPriceConfigs map[string]ModelConfigLocal) error {
	if modelPriceConfigs == nil || len(modelPriceConfigs) == 0 {
		channel.ModelConfigs = nil
		return nil
	}

	// Validate the configurations before setting
	if err := channel.validateModelPriceConfigs(modelPriceConfigs); err != nil {
		return fmt.Errorf("invalid model price configurations: %w", err)
	}

	jsonBytes, err := json.Marshal(modelPriceConfigs)
	if err != nil {
		return fmt.Errorf("failed to marshal model price configurations: %w", err)
	}

	jsonStr := string(jsonBytes)
	channel.ModelConfigs = &jsonStr
	return nil
}

// GetModelPriceConfig returns the price configuration for a specific model
func (channel *Channel) GetModelPriceConfig(modelName string) *ModelConfigLocal {
	configs := channel.GetModelPriceConfigs()
	if configs == nil {
		return nil
	}

	if config, exists := configs[modelName]; exists {
		return &config
	}

	return nil
}

// GetModelRatioFromConfigs extracts model ratios from the unified ModelConfigs
func (channel *Channel) GetModelRatioFromConfigs() map[string]float64 {
	configs := channel.GetModelPriceConfigs()
	if configs == nil {
		return nil
	}

	modelRatios := make(map[string]float64)
	for modelName, config := range configs {
		if config.Ratio != 0 {
			modelRatios[modelName] = config.Ratio
		}
	}

	if len(modelRatios) == 0 {
		return nil
	}

	return modelRatios
}

// GetCompletionRatioFromConfigs extracts completion ratios from the unified ModelConfigs
func (channel *Channel) GetCompletionRatioFromConfigs() map[string]float64 {
	configs := channel.GetModelPriceConfigs()
	if configs == nil {
		return nil
	}

	completionRatios := make(map[string]float64)
	for modelName, config := range configs {
		if config.CompletionRatio != 0 {
			completionRatios[modelName] = config.CompletionRatio
		}
	}

	if len(completionRatios) == 0 {
		return nil
	}

	return completionRatios
}

func (channel *Channel) GetInferenceProfileArnMap() map[string]string {
	if channel.InferenceProfileArnMap == nil || *channel.InferenceProfileArnMap == "" || *channel.InferenceProfileArnMap == "{}" {
		return nil
	}
	arnMap := make(map[string]string)
	err := json.Unmarshal([]byte(*channel.InferenceProfileArnMap), &arnMap)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal inference profile ARN map for channel %d, error: %s", channel.Id, err.Error()))
		return nil
	}
	return arnMap
}

func (channel *Channel) SetInferenceProfileArnMap(arnMap map[string]string) error {
	if arnMap == nil || len(arnMap) == 0 {
		channel.InferenceProfileArnMap = nil
		return nil
	}

	// Validate that keys and values are not empty
	for key, value := range arnMap {
		if key == "" || value == "" {
			return fmt.Errorf("inference profile ARN map cannot contain empty keys or values")
		}
	}

	jsonBytes, err := json.Marshal(arnMap)
	if err != nil {
		return err
	}
	jsonStr := string(jsonBytes)
	channel.InferenceProfileArnMap = &jsonStr
	return nil
}

// ValidateInferenceProfileArnMapJSON validates a JSON string for inference profile ARN mapping
func ValidateInferenceProfileArnMapJSON(jsonStr string) error {
	if jsonStr == "" {
		return nil // Empty is allowed
	}

	var arnMap map[string]string
	err := json.Unmarshal([]byte(jsonStr), &arnMap)
	if err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	// Validate that keys and values are not empty
	for key, value := range arnMap {
		if key == "" {
			return fmt.Errorf("inference profile ARN map cannot contain empty keys")
		}
		if value == "" {
			return fmt.Errorf("inference profile ARN map cannot contain empty values")
		}
	}

	return nil
}

func (channel *Channel) Insert() error {
	var err error
	err = DB.Create(channel).Error
	if err != nil {
		return err
	}
	err = channel.AddAbilities()
	if err == nil {
		InitChannelCache()
	}
	return err
}

func (channel *Channel) Update() error {
	var err error
	err = DB.Model(channel).Updates(channel).Error
	if err != nil {
		return err
	}
	DB.Model(channel).First(channel, "id = ?", channel.Id)
	err = channel.UpdateAbilities()
	if err == nil {
		InitChannelCache()
	}
	return err
}

func (channel *Channel) UpdateResponseTime(responseTime int64) {
	err := DB.Model(channel).Select("response_time", "test_time").Updates(Channel{
		TestTime:     helper.GetTimestamp(),
		ResponseTime: int(responseTime),
	}).Error
	if err != nil {
		logger.SysError("failed to update response time: " + err.Error())
	}
}

func (channel *Channel) UpdateBalance(balance float64) {
	err := DB.Model(channel).Select("balance_updated_time", "balance").Updates(Channel{
		BalanceUpdatedTime: helper.GetTimestamp(),
		Balance:            balance,
	}).Error
	if err != nil {
		logger.SysError("failed to update balance: " + err.Error())
	}
}

func (channel *Channel) Delete() error {
	var err error
	err = DB.Delete(channel).Error
	if err != nil {
		return err
	}
	err = channel.DeleteAbilities()
	if err == nil {
		InitChannelCache()
	}
	return err
}

func (channel *Channel) LoadConfig() (ChannelConfig, error) {
	var cfg ChannelConfig
	if channel.Config == "" {
		return cfg, nil
	}
	err := json.Unmarshal([]byte(channel.Config), &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// GetModelRatio returns the channel-specific model ratio map
// DEPRECATED: Use GetModelPriceConfigs() instead. This method is kept for backward compatibility.
func (channel *Channel) GetModelRatio() map[string]float64 {
	if channel.ModelRatio == nil || *channel.ModelRatio == "" || *channel.ModelRatio == "{}" {
		return nil
	}
	modelRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(*channel.ModelRatio), &modelRatio)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal model ratio for channel %d, error: %s", channel.Id, err.Error()))
		return nil
	}
	return modelRatio
}

// GetCompletionRatio returns the channel-specific completion ratio map
// DEPRECATED: Use GetModelPriceConfigs() instead. This method is kept for backward compatibility.
func (channel *Channel) GetCompletionRatio() map[string]float64 {
	if channel.CompletionRatio == nil || *channel.CompletionRatio == "" || *channel.CompletionRatio == "{}" {
		return nil
	}
	completionRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(*channel.CompletionRatio), &completionRatio)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to unmarshal completion ratio for channel %d, error: %s", channel.Id, err.Error()))
		return nil
	}
	return completionRatio
}

// SetModelRatio sets the channel-specific model ratio map
// DEPRECATED: Use SetModelPriceConfigs() instead. This method is kept for backward compatibility.
func (channel *Channel) SetModelRatio(modelRatio map[string]float64) error {
	if modelRatio == nil || len(modelRatio) == 0 {
		channel.ModelRatio = nil
		return nil
	}
	jsonBytes, err := json.Marshal(modelRatio)
	if err != nil {
		return err
	}
	jsonStr := string(jsonBytes)
	channel.ModelRatio = &jsonStr
	return nil
}

// SetCompletionRatio sets the channel-specific completion ratio map
// DEPRECATED: Use SetModelPriceConfigs() instead. This method is kept for backward compatibility.
func (channel *Channel) SetCompletionRatio(completionRatio map[string]float64) error {
	if completionRatio == nil || len(completionRatio) == 0 {
		channel.CompletionRatio = nil
		return nil
	}
	jsonBytes, err := json.Marshal(completionRatio)
	if err != nil {
		return err
	}
	jsonStr := string(jsonBytes)
	channel.CompletionRatio = &jsonStr
	return nil
}

func UpdateChannelStatusById(id int, status int) {
	err := UpdateAbilityStatus(id, status == ChannelStatusEnabled)
	if err != nil {
		logger.SysError("failed to update ability status: " + err.Error())
	}
	err = DB.Model(&Channel{}).Where("id = ?", id).Update("status", status).Error
	if err != nil {
		logger.SysError("failed to update channel status: " + err.Error())
	}
	if err == nil {
		InitChannelCache()
	}
}

func UpdateChannelUsedQuota(id int, quota int64) {
	if config.BatchUpdateEnabled {
		addNewRecord(BatchUpdateTypeChannelUsedQuota, id, quota)
		return
	}
	updateChannelUsedQuota(id, quota)
}

func updateChannelUsedQuota(id int, quota int64) {
	err := DB.Model(&Channel{}).Where("id = ?", id).Update("used_quota", gorm.Expr("used_quota + ?", quota)).Error
	if err != nil {
		logger.SysError("failed to update channel used quota: " + err.Error())
	}
}

func DeleteChannelByStatus(status int64) (int64, error) {
	result := DB.Where("status = ?", status).Delete(&Channel{})
	if result.Error == nil {
		InitChannelCache()
	}
	return result.RowsAffected, result.Error
}

func DeleteDisabledChannel() (int64, error) {
	result := DB.Where("status = ? or status = ?", ChannelStatusAutoDisabled, ChannelStatusManuallyDisabled).Delete(&Channel{})
	if result.Error == nil {
		InitChannelCache()
	}
	return result.RowsAffected, result.Error
}

// MigrateHistoricalPricingToModelConfigs migrates historical ModelRatio and CompletionRatio data
// into the new unified ModelConfigs format for a single channel
func (channel *Channel) MigrateHistoricalPricingToModelConfigs() error {
	// Validate channel
	if channel == nil {
		return fmt.Errorf("channel is nil")
	}

	// Get existing ModelRatio and CompletionRatio data with validation
	var modelRatios map[string]float64
	var completionRatios map[string]float64
	var migrationErrors []string

	// Safely get ModelRatio
	if channel.ModelRatio != nil && *channel.ModelRatio != "" && *channel.ModelRatio != "{}" {
		if err := json.Unmarshal([]byte(*channel.ModelRatio), &modelRatios); err != nil {
			migrationErrors = append(migrationErrors, fmt.Sprintf("invalid ModelRatio JSON: %s", err.Error()))
		} else {
			// Validate ModelRatio values
			for modelName, ratio := range modelRatios {
				if modelName == "" {
					migrationErrors = append(migrationErrors, "empty model name in ModelRatio")
				}
				if ratio < 0 {
					migrationErrors = append(migrationErrors, fmt.Sprintf("negative ratio for model %s: %f", modelName, ratio))
				}
			}
		}
	}

	// Safely get CompletionRatio
	if channel.CompletionRatio != nil && *channel.CompletionRatio != "" && *channel.CompletionRatio != "{}" {
		if err := json.Unmarshal([]byte(*channel.CompletionRatio), &completionRatios); err != nil {
			migrationErrors = append(migrationErrors, fmt.Sprintf("invalid CompletionRatio JSON: %s", err.Error()))
		} else {
			// Validate CompletionRatio values
			for modelName, ratio := range completionRatios {
				if modelName == "" {
					migrationErrors = append(migrationErrors, "empty model name in CompletionRatio")
				}
				if ratio < 0 {
					migrationErrors = append(migrationErrors, fmt.Sprintf("negative completion ratio for model %s: %f", modelName, ratio))
				}
			}
		}
	}

	// Report validation errors but continue with valid data
	if len(migrationErrors) > 0 {
		logger.SysError(fmt.Sprintf("Channel %d has validation errors in historical data: %v", channel.Id, migrationErrors))
		// Don't return error - continue with valid data
	}

	// Skip if no valid historical data to migrate
	if (modelRatios == nil || len(modelRatios) == 0) && (completionRatios == nil || len(completionRatios) == 0) {
		return nil
	}

	// Check if ModelConfigs already has unified data
	existingConfigs := channel.GetModelPriceConfigs()
	if existingConfigs != nil && len(existingConfigs) > 0 {
		// Check if existing configs have pricing data (not just MaxTokens)
		hasPricingData := false
		for _, config := range existingConfigs {
			if config.Ratio != 0 || config.CompletionRatio != 0 {
				hasPricingData = true
				break
			}
		}

		if hasPricingData {
			logger.SysLog(fmt.Sprintf("Channel %d already has pricing data in ModelConfigs, skipping historical migration", channel.Id))
			return nil
		}

		// Merge historical pricing with existing MaxTokens data
		logger.SysLog(fmt.Sprintf("Channel %d has MaxTokens data, merging with historical pricing", channel.Id))
	} else {
		existingConfigs = make(map[string]ModelConfigLocal)
	}

	// Collect all valid model names from both ratios and existing configs
	allModelNames := make(map[string]bool)
	if modelRatios != nil {
		for modelName, ratio := range modelRatios {
			// Skip invalid entries
			if modelName != "" && ratio >= 0 {
				allModelNames[modelName] = true
			}
		}
	}
	if completionRatios != nil {
		for modelName, ratio := range completionRatios {
			// Skip invalid entries
			if modelName != "" && ratio >= 0 {
				allModelNames[modelName] = true
			}
		}
	}
	for modelName := range existingConfigs {
		if modelName != "" {
			allModelNames[modelName] = true
		}
	}

	// Create unified ModelConfigs from all data sources
	modelConfigs := make(map[string]ModelConfigLocal)
	for modelName := range allModelNames {
		config := ModelConfigLocal{}

		// Start with existing config if available
		if existingConfig, exists := existingConfigs[modelName]; exists {
			config = existingConfig
		}

		// Add/override pricing data from historical sources (only valid data)
		if modelRatios != nil {
			if ratio, exists := modelRatios[modelName]; exists && ratio >= 0 {
				config.Ratio = ratio
			}
		}

		if completionRatios != nil {
			if completionRatio, exists := completionRatios[modelName]; exists && completionRatio >= 0 {
				config.CompletionRatio = completionRatio
			}
		}

		// Add if we have any data (pricing or MaxTokens)
		if config.Ratio != 0 || config.CompletionRatio != 0 || config.MaxTokens != 0 {
			modelConfigs[modelName] = config
		}
	}

	// Save the migrated data to ModelConfigs
	if len(modelConfigs) > 0 {
		// Log the models being migrated for debugging
		var modelNames []string
		for modelName := range modelConfigs {
			modelNames = append(modelNames, modelName)
		}
		logger.SysLog(fmt.Sprintf("Channel %d (type %d) migrating models: %v", channel.Id, channel.Type, modelNames))

		err := channel.SetModelPriceConfigs(modelConfigs)
		if err != nil {
			logger.SysError(fmt.Sprintf("Failed to set migrated ModelConfigs for channel %d: %s", channel.Id, err.Error()))
			return err
		}

		logger.SysLog(fmt.Sprintf("Successfully migrated historical pricing data to ModelConfigs for channel %d (%d models)", channel.Id, len(modelConfigs)))
	}

	return nil
}

// MigrateChannelFieldsToText migrates ModelConfigs and ModelMapping fields from varchar(1024) to text type.
//
// Background:
// The original varchar(1024) length was insufficient for complex model configurations, especially when:
// - Multiple models are configured with detailed pricing information (ratio, completion_ratio, max_tokens)
// - Long model names or complex mapping values are used
// - Channel-specific configurations grow beyond the 1024 character limit
//
// This migration is essential because:
// 1. Modern AI models have longer names and more complex configurations
// 2. Users need to configure pricing for dozens of models per channel
// 3. JSON serialization of comprehensive model configs easily exceeds 1024 chars
// 4. Truncated configurations lead to data loss and system errors
//
// The migration is designed to be:
// - Idempotent: Can be run multiple times safely
// - Database-agnostic: Supports MySQL, PostgreSQL, and SQLite
// - Data-preserving: All existing data is maintained during the migration
// - Transaction-safe: Uses database transactions to ensure data integrity
//
// This function should be called during application startup before any channel operations.
func MigrateChannelFieldsToText() error {
	logger.SysLog("Starting migration of ModelConfigs and ModelMapping fields to TEXT type")

	// Check if migration is needed by examining current schema (idempotency check)
	// This ensures the migration can be run multiple times safely
	needsMigration, err := checkIfFieldMigrationNeeded()
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to check if field migration is needed: %s", err.Error()))
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if !needsMigration {
		logger.SysLog("ModelConfigs and ModelMapping fields are already TEXT type - no migration needed")
		return nil
	}

	logger.SysLog("Column type migration required - proceeding with migration")

	// Perform the actual migration with proper transaction handling
	return performFieldMigration()
}

// performFieldMigration executes the actual database schema changes to migrate fields to TEXT type.
// This function uses database transactions to ensure data integrity and provides detailed error handling.
func performFieldMigration() error {
	// Use transaction for data integrity - ensures all-or-nothing migration
	tx := DB.Begin()
	if tx.Error != nil {
		logger.SysError(fmt.Sprintf("Failed to start column migration transaction: %s", tx.Error.Error()))
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	// Ensure transaction is properly handled in case of panic or error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.SysError(fmt.Sprintf("Column migration panicked, rolled back: %v", r))
		}
	}()

	// Perform database-specific column type changes
	var err error
	if common.UsingMySQL {
		err = performMySQLFieldMigration(tx)
	} else if common.UsingPostgreSQL {
		err = performPostgreSQLFieldMigration(tx)
	} else {
		// This should not happen due to the check in checkIfFieldMigrationNeeded,
		// but we handle it for safety
		tx.Rollback()
		return fmt.Errorf("unsupported database type for field migration")
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		logger.SysError(fmt.Sprintf("Failed to commit column migration transaction: %s", err.Error()))
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	logger.SysLog("Successfully migrated ModelConfigs and ModelMapping columns to TEXT type")
	return nil
}

// performMySQLFieldMigration performs the MySQL-specific column type migration.
func performMySQLFieldMigration(tx *gorm.DB) error {
	logger.SysLog("Performing MySQL field migration")

	// MySQL: Use MODIFY COLUMN to change type while preserving data
	err := tx.Exec("ALTER TABLE channels MODIFY COLUMN model_configs TEXT DEFAULT ''").Error
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to migrate model_configs column in MySQL: %s", err.Error()))
		return fmt.Errorf("failed to migrate model_configs column: %w", err)
	}

	err = tx.Exec("ALTER TABLE channels MODIFY COLUMN model_mapping TEXT DEFAULT ''").Error
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to migrate model_mapping column in MySQL: %s", err.Error()))
		return fmt.Errorf("failed to migrate model_mapping column: %w", err)
	}

	logger.SysLog("MySQL field migration completed successfully")
	return nil
}

// performPostgreSQLFieldMigration performs the PostgreSQL-specific column type migration.
func performPostgreSQLFieldMigration(tx *gorm.DB) error {
	logger.SysLog("Performing PostgreSQL field migration")

	// PostgreSQL: Use ALTER COLUMN TYPE to change column type
	err := tx.Exec("ALTER TABLE channels ALTER COLUMN model_configs TYPE TEXT").Error
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to migrate model_configs column in PostgreSQL: %s", err.Error()))
		return fmt.Errorf("failed to migrate model_configs column: %w", err)
	}

	err = tx.Exec("ALTER TABLE channels ALTER COLUMN model_mapping TYPE TEXT").Error
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to migrate model_mapping column in PostgreSQL: %s", err.Error()))
		return fmt.Errorf("failed to migrate model_mapping column: %w", err)
	}

	logger.SysLog("PostgreSQL field migration completed successfully")
	return nil
}

// checkIfFieldMigrationNeeded checks if ModelConfigs and ModelMapping fields need to be migrated to TEXT type.
// This function provides idempotency by checking the current column types in the database.
// Returns true if migration is needed, false if fields are already TEXT type.
func checkIfFieldMigrationNeeded() (bool, error) {
	if common.UsingMySQL {
		// Check MySQL column types for both fields
		var modelConfigsType, modelMappingType string

		// Check model_configs column type
		err := DB.Raw(`SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'channels' AND COLUMN_NAME = 'model_configs'`).
			Scan(&modelConfigsType).Error
		if err != nil {
			return false, fmt.Errorf("failed to check model_configs column type in MySQL: %w", err)
		}

		// Check model_mapping column type
		err = DB.Raw(`SELECT DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'channels' AND COLUMN_NAME = 'model_mapping'`).
			Scan(&modelMappingType).Error
		if err != nil {
			return false, fmt.Errorf("failed to check model_mapping column type in MySQL: %w", err)
		}

		// Migration needed if either field is still varchar
		return modelConfigsType == "varchar" || modelMappingType == "varchar", nil

	} else if common.UsingPostgreSQL {
		// Check PostgreSQL column types for both fields
		var modelConfigsType, modelMappingType string

		// Check model_configs column type
		err := DB.Raw(`SELECT data_type FROM information_schema.columns
			WHERE table_name = 'channels' AND column_name = 'model_configs'`).
			Scan(&modelConfigsType).Error
		if err != nil {
			return false, fmt.Errorf("failed to check model_configs column type in PostgreSQL: %w", err)
		}

		// Check model_mapping column type
		err = DB.Raw(`SELECT data_type FROM information_schema.columns
			WHERE table_name = 'channels' AND column_name = 'model_mapping'`).
			Scan(&modelMappingType).Error
		if err != nil {
			return false, fmt.Errorf("failed to check model_mapping column type in PostgreSQL: %w", err)
		}

		// Migration needed if either field is still character varying (varchar)
		return modelConfigsType == "character varying" || modelMappingType == "character varying", nil

	} else if common.UsingSQLite {
		// SQLite is flexible with column types and doesn't enforce strict typing
		// TEXT and VARCHAR are treated the same way, so no migration is needed
		logger.SysLog("SQLite detected - column type migration not required (SQLite is flexible with text types)")
		return false, nil

	} else {
		// Unknown database type - assume no migration needed to be safe
		logger.SysLog("Unknown database type detected - skipping column type migration")
		return false, nil
	}
}

// MigrateAllChannelModelConfigs migrates all channels' ModelConfigs from old format to new format
// and also migrates historical ModelRatio/CompletionRatio data to the new unified format
// This should be called during application startup to handle existing data
func MigrateAllChannelModelConfigs() error {
	logger.SysLog("Starting migration of all channel ModelConfigs and historical pricing data")

	var channels []*Channel
	err := DB.Find(&channels).Error
	if err != nil {
		logger.SysError(fmt.Sprintf("Failed to fetch channels for ModelConfigs migration: %s", err.Error()))
		return fmt.Errorf("failed to fetch channels: %w", err)
	}

	if len(channels) == 0 {
		logger.SysLog("No channels found for migration")
		return nil
	}

	migratedCount := 0
	historicalMigratedCount := 0
	errorCount := 0
	var migrationErrors []string

	// Use transaction for data integrity
	tx := DB.Begin()
	if tx.Error != nil {
		logger.SysError(fmt.Sprintf("Failed to start migration transaction: %s", tx.Error.Error()))
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.SysError(fmt.Sprintf("Migration panicked, rolled back: %v", r))
		}
	}()

	for _, channel := range channels {
		channelUpdated := false
		originalModelConfigs := ""
		if channel.ModelConfigs != nil {
			originalModelConfigs = *channel.ModelConfigs
		}

		// First, migrate existing ModelConfigs from old format to new format (PR format -> unified format)
		if channel.ModelConfigs != nil && *channel.ModelConfigs != "" && *channel.ModelConfigs != "{}" {
			err := channel.MigrateModelConfigsToModelPrice()
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to migrate ModelConfigs for channel %d: %s", channel.Id, err.Error())
				logger.SysError(errorMsg)
				migrationErrors = append(migrationErrors, errorMsg)
				errorCount++
				continue
			}
			channelUpdated = true
			migratedCount++
		}

		// Second, migrate historical ModelRatio/CompletionRatio data to ModelConfigs
		err := channel.MigrateHistoricalPricingToModelConfigs()
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to migrate historical pricing for channel %d: %s", channel.Id, err.Error())
			logger.SysError(errorMsg)
			migrationErrors = append(migrationErrors, errorMsg)
			errorCount++
			continue
		}

		// Check if historical migration actually created ModelConfigs data
		if channel.ModelConfigs != nil && *channel.ModelConfigs != "" && *channel.ModelConfigs != "{}" {
			if !channelUpdated { // Only count if it wasn't already counted in the first migration
				historicalMigratedCount++
				channelUpdated = true
			}
		}

		// Save the migrated channel back to database if any changes were made
		if channelUpdated {
			// Validate the final result before saving
			finalConfigs := channel.GetModelPriceConfigs()
			if err := channel.validateModelPriceConfigs(finalConfigs); err != nil {
				errorMsg := fmt.Sprintf("Migration validation failed for channel %d: %s", channel.Id, err.Error())
				logger.SysError(errorMsg)
				migrationErrors = append(migrationErrors, errorMsg)
				errorCount++
				// Restore original data
				if originalModelConfigs != "" {
					channel.ModelConfigs = &originalModelConfigs
				} else {
					channel.ModelConfigs = nil
				}
				continue
			}

			err = tx.Model(channel).Update("model_configs", channel.ModelConfigs).Error
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to save migrated ModelConfigs for channel %d: %s", channel.Id, err.Error())
				logger.SysError(errorMsg)
				migrationErrors = append(migrationErrors, errorMsg)
				errorCount++
				continue
			}
		}
	}

	// Commit transaction if no critical errors
	if errorCount == 0 {
		if err := tx.Commit().Error; err != nil {
			logger.SysError(fmt.Sprintf("Failed to commit migration transaction: %s", err.Error()))
			return fmt.Errorf("failed to commit migration: %w", err)
		}
	} else {
		tx.Rollback()
		logger.SysError(fmt.Sprintf("Migration had %d errors, rolled back transaction", errorCount))

		// If more than 50% of channels failed, return error to prevent silent data loss
		failureRate := float64(errorCount) / float64(len(channels))
		if failureRate > 0.5 {
			return fmt.Errorf("migration failed for %d/%d channels (%.1f%%), rolled back to prevent data loss",
				errorCount, len(channels), failureRate*100)
		}

		// For lower failure rates, log errors but don't fail startup
		logger.SysError(fmt.Sprintf("Migration completed with %d errors out of %d channels", errorCount, len(channels)))
	}

	// Log final results
	if migratedCount > 0 {
		logger.SysLog(fmt.Sprintf("Successfully migrated ModelConfigs format for %d channels", migratedCount))
	}
	if historicalMigratedCount > 0 {
		logger.SysLog(fmt.Sprintf("Successfully migrated historical pricing data for %d channels", historicalMigratedCount))
	}
	if errorCount > 0 {
		logger.SysError(fmt.Sprintf("Migration completed with %d errors", errorCount))
		for _, errMsg := range migrationErrors {
			logger.SysError(fmt.Sprintf("Migration error: %s", errMsg))
		}
	}
	if migratedCount == 0 && historicalMigratedCount == 0 && errorCount == 0 {
		logger.SysLog("No channels required data migration")
	}

	return nil
}
