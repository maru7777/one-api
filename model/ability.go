package model

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	gutils "github.com/Laisky/go-utils/v5"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/utils"
)

type Ability struct {
	Group        string     `json:"group" gorm:"type:varchar(32);primaryKey;autoIncrement:false"`
	Model        string     `json:"model" gorm:"primaryKey;autoIncrement:false"`
	ChannelId    int        `json:"channel_id" gorm:"primaryKey;autoIncrement:false;index"`
	Enabled      bool       `json:"enabled"`
	Priority     *int64     `json:"priority" gorm:"bigint;default:0;index"`
	SuspendUntil *time.Time `json:"suspend_until,omitempty" gorm:"index"`
}

func GetRandomSatisfiedChannel(group string, model string, ignoreFirstPriority bool) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	now := time.Now()

	var err error = nil
	var channelQuery *gorm.DB
	if ignoreFirstPriority {
		channelQuery = DB.Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND (suspend_until IS NULL OR suspend_until < ?)", group, model, now)
	} else {
		maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND (suspend_until IS NULL OR suspend_until < ?)", group, model, now)
		channelQuery = DB.Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND priority = (?) AND (suspend_until IS NULL OR suspend_until < ?)", group, model, maxPrioritySubQuery, now)
	}
	if common.UsingSQLite || common.UsingPostgreSQL {
		err = channelQuery.Order("RANDOM()").First(&ability).Error
	} else {
		err = channelQuery.Order("RAND()").First(&ability).Error
	}
	if err != nil {
		return nil, errors.Wrap(err, "get random satisfied channel")
	}

	channel := Channel{}
	channel.Id = ability.ChannelId
	err = DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

func (channel *Channel) AddAbilities() error {
	models_ := strings.Split(channel.Models, ",")
	models_ = utils.DeDuplication(models_)
	groups_ := strings.Split(channel.Group, ",")
	abilities := make([]Ability, 0, len(models_))
	for _, model := range models_ {
		for _, group := range groups_ {
			ability := Ability{
				Group:        group,
				Model:        model,
				ChannelId:    channel.Id,
				Enabled:      channel.Status == ChannelStatusEnabled,
				Priority:     channel.Priority,
				SuspendUntil: nil, // Explicitly nil on new creation
			}
			abilities = append(abilities, ability)
		}
	}
	return DB.Create(&abilities).Error
}

func (channel *Channel) DeleteAbilities() error {
	return DB.Where("channel_id = ?", channel.Id).Delete(&Ability{}).Error
}

// UpdateAbilities updates abilities of this channel.
// Make sure the channel is completed before calling this function.
func (channel *Channel) UpdateAbilities() error {
	// A quick and dirty way to update abilities
	// First delete all abilities of this channel
	err := channel.DeleteAbilities()
	if err != nil {
		return err
	}
	// Then add new abilities
	err = channel.AddAbilities()
	if err != nil {
		return err
	}
	return nil
}

func UpdateAbilityStatus(channelId int, status bool) error {
	return DB.Model(&Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func GetGroupModels(ctx context.Context, group string) ([]string, error) {
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	var models []string
	err := DB.Model(&Ability{}).Distinct("model").Where(groupCol+" = ? AND enabled = "+trueVal+" AND (suspend_until IS NULL OR suspend_until < ?)", group, time.Now()).Pluck("model", &models).Error
	if err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, err
}

var getGroupModelsV2Cache = gutils.NewExpCache[[]EnabledAbility](context.Background(), time.Second*10)

type EnabledAbility struct {
	Model       string `json:"model" gorm:"model"`
	ChannelType int    `json:"channel_type" gorm:"channel_type"`
}

// GetGroupModelsV2 returns all enabled models for this group with their channel names.
func GetGroupModelsV2(ctx context.Context, group string) ([]EnabledAbility, error) {
	// get from cache first
	if models, ok := getGroupModelsV2Cache.Load(group); ok {
		return models, nil
	}

	// prepare query based on database type
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	now := time.Now()

	// query with JOIN to get model and channel name in a single query
	var models []EnabledAbility
	query := DB.Model(&Ability{}).
		Select("DISTINCT abilities.model AS model, channels.type AS channel_type").
		Joins("JOIN channels ON abilities.channel_id = channels.id").
		Where("abilities."+groupCol+" = ? AND abilities.enabled = "+trueVal+" AND (abilities.suspend_until IS NULL OR abilities.suspend_until < ?)", group, now).
		Order("abilities.model")

	err := query.Find(&models).Error
	if err != nil {
		return nil, errors.Wrap(err, "get group models")
	}

	// store in cache
	getGroupModelsV2Cache.Store(group, models)

	return models, nil
}

// SuspendAbility sets the SuspendUntil timestamp for a given ability.
func SuspendAbility(ctx context.Context, group string, modelName string, channelId int, duration time.Duration) error {
	if group == "" || modelName == "" || channelId == 0 {
		return errors.New("group, modelName, and channelId must be specified for suspending ability")
	}
	suspendTime := time.Now().Add(duration)

	// Handle database-specific identifier quoting like other functions
	groupCol := "`group`"
	modelCol := "`model`"
	channelCol := "`channel_id`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		modelCol = `"model"`
		channelCol = `"channel_id"`
	}

	return DB.Model(&Ability{}).
		Where(groupCol+" = ? AND "+modelCol+" = ? AND "+channelCol+" = ?",
			group, modelName, channelId).
		Update("suspend_until", suspendTime).Error
}

func GetRandomSatisfiedChannelExcluding(group string, model string, ignoreFirstPriority bool, excludeChannelIds map[int]bool) (*Channel, error) {
	ability := Ability{}
	groupCol := "`group`"
	trueVal := "1"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		trueVal = "true"
	}
	now := time.Now()

	var err error = nil
	var channelQuery *gorm.DB

	// Build base query with exclusions
	baseCondition := groupCol + " = ? AND model = ? AND enabled = " + trueVal + " AND (suspend_until IS NULL OR suspend_until < ?)"
	if len(excludeChannelIds) > 0 {
		var excludeIds []int
		for channelId := range excludeChannelIds {
			excludeIds = append(excludeIds, channelId)
		}
		baseCondition += " AND channel_id NOT IN (?)"
	}

	if ignoreFirstPriority {
		// For ignoreFirstPriority=true, we want to select from lower priority channels
		// First, find the maximum priority among available channels (excluding failed ones)
		maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND (suspend_until IS NULL OR suspend_until < ?)", group, model, now)
		if len(excludeChannelIds) > 0 {
			var excludeIds []int
			for channelId := range excludeChannelIds {
				excludeIds = append(excludeIds, channelId)
			}
			maxPrioritySubQuery = maxPrioritySubQuery.Where("channel_id NOT IN (?)", excludeIds)
		}

		// Then select channels with priority less than the maximum priority
		if len(excludeChannelIds) > 0 {
			var excludeIds []int
			for channelId := range excludeChannelIds {
				excludeIds = append(excludeIds, channelId)
			}
			channelQuery = DB.Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND priority < (?) AND (suspend_until IS NULL OR suspend_until < ?) AND channel_id NOT IN (?)", group, model, maxPrioritySubQuery, now, excludeIds)
		} else {
			channelQuery = DB.Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND priority < (?) AND (suspend_until IS NULL OR suspend_until < ?)", group, model, maxPrioritySubQuery, now)
		}
	} else {
		// For ignoreFirstPriority=false, select from highest priority channels
		// First check if there are any available channels after exclusions
		var availableCount int64
		countQuery := DB.Model(&Ability{}).Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND (suspend_until IS NULL OR suspend_until < ?)", group, model, now)
		if len(excludeChannelIds) > 0 {
			var excludeIds []int
			for channelId := range excludeChannelIds {
				excludeIds = append(excludeIds, channelId)
			}
			countQuery = countQuery.Where("channel_id NOT IN (?)", excludeIds)
		}
		countQuery.Count(&availableCount)

		if availableCount == 0 {
			// Simple error message to avoid performance overhead
			// Detailed diagnostics can be enabled via debug mode if needed
			var excludeIds []int
			for channelId := range excludeChannelIds {
				excludeIds = append(excludeIds, channelId)
			}

			errorMsg := fmt.Sprintf("no channels available for model %s in group %s after excluding %d channels",
				model, group, len(excludeIds))
			return nil, errors.New(errorMsg)
		}

		// Now find the maximum priority among available channels
		maxPrioritySubQuery := DB.Model(&Ability{}).Select("MAX(priority)").Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND (suspend_until IS NULL OR suspend_until < ?)", group, model, now)
		if len(excludeChannelIds) > 0 {
			var excludeIds []int
			for channelId := range excludeChannelIds {
				excludeIds = append(excludeIds, channelId)
			}
			// Add exclusion to the subquery as well
			maxPrioritySubQuery = maxPrioritySubQuery.Where("channel_id NOT IN (?)", excludeIds)
			channelQuery = DB.Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND priority = (?) AND (suspend_until IS NULL OR suspend_until < ?) AND channel_id NOT IN (?)", group, model, maxPrioritySubQuery, now, excludeIds)
		} else {
			channelQuery = DB.Where(groupCol+" = ? AND model = ? AND enabled = "+trueVal+" AND priority = (?) AND (suspend_until IS NULL OR suspend_until < ?)", group, model, maxPrioritySubQuery, now)
		}
	}

	if common.UsingSQLite || common.UsingPostgreSQL {
		err = channelQuery.Order("RANDOM()").First(&ability).Error
	} else {
		err = channelQuery.Order("RAND()").First(&ability).Error
	}
	if err != nil {
		return nil, errors.Wrap(err, "get random satisfied channel excluding failed ones")
	}

	channel := Channel{}
	channel.Id = ability.ChannelId
	err = DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}
