package model

import (
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/pkg/errors"
)

type UserRequestCost struct {
	Id          int     `json:"id"`
	CreatedTime int64   `json:"created_time" gorm:"bigint"`
	UserID      int     `json:"user_id"`
	RequestID   string  `json:"request_id"`
	Quota       int64   `json:"quota"`
	CostUSD     float64 `json:"cost_usd" gorm:"-"`
}

// NewUserRequestCost create a new UserRequestCost
func NewUserRequestCost(userID int, quotaID string, quota int64) *UserRequestCost {
	return &UserRequestCost{
		CreatedTime: helper.GetTimestamp(),
		UserID:      userID,
		RequestID:   quotaID,
		Quota:       quota,
	}
}

func (docu *UserRequestCost) Insert() error {
	var err error
	err = DB.Create(docu).Error
	return errors.Wrap(err, "failed to insert UserRequestCost")
}

// GetCostByRequestId get cost by request id
func GetCostByRequestId(reqid string) (*UserRequestCost, error) {
	if reqid == "" {
		return nil, errors.New("request id is empty")
	}

	docu := &UserRequestCost{RequestID: reqid}
	var err error = nil
	if err = DB.First(docu, "request_id = ?", reqid).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get cost by request id")
	}

	docu.CostUSD = float64(docu.Quota) / 500000
	return docu, nil
}
