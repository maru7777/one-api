package model

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

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
	Group              string  `json:"group" gorm:"type:varchar(32);default:'default'"`
	UsedQuota          int64   `json:"used_quota" gorm:"bigint;default:0"`
	ModelMapping       *string `json:"model_mapping" gorm:"type:varchar(1024);default:''"`
	Priority           *int64  `json:"priority" gorm:"bigint;default:0"`
	Config             string  `json:"config"`
	SystemPrompt       *string `json:"system_prompt" gorm:"type:text"`
	RateLimit          *int    `json:"ratelimit" gorm:"column:ratelimit;default:0"`
	// Channel-specific pricing tables
	ModelRatio      *string `json:"model_ratio" gorm:"type:text"`      // JSON string of model pricing ratios
	CompletionRatio *string `json:"completion_ratio" gorm:"type:text"` // JSON string of completion pricing ratios
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
