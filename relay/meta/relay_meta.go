package meta

import (
	"github.com/Laisky/one-api/common/config"
	"github.com/Laisky/one-api/relay/adaptor/azure"
	"github.com/Laisky/one-api/relay/channeltype"
	"github.com/Laisky/one-api/relay/relaymode"
	"github.com/gin-gonic/gin"
	"strings"
)

type Meta struct {
	Mode            int
	ChannelType     int
	ChannelId       int
	TokenId         int
	TokenName       string
	UserId          int
	Group           string
	ModelMapping    map[string]string
	BaseURL         string
	APIVersion      string
	APIKey          string
	APIType         int
	Config          map[string]string
	IsStream        bool
	OriginModelName string
	ActualModelName string
	RequestURLPath  string
	PromptTokens    int // only for DoResponse
	ChannelRatio    float64
}

func GetByContext(c *gin.Context) *Meta {
	meta := Meta{
		Mode:           relaymode.GetByPath(c.Request.URL.Path),
		ChannelType:    c.GetInt("channel"),
		ChannelId:      c.GetInt("channel_id"),
		TokenId:        c.GetInt("token_id"),
		TokenName:      c.GetString("token_name"),
		UserId:         c.GetInt("id"),
		Group:          c.GetString("group"),
		ModelMapping:   c.GetStringMapString("model_mapping"),
		BaseURL:        c.GetString("base_url"),
		APIVersion:     c.GetString(config.KeyAPIVersion),
		APIKey:         strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		Config:         nil,
		RequestURLPath: c.Request.URL.String(),
		ChannelRatio:   c.GetFloat64("channel_ratio"),
	}
	if meta.ChannelType == channeltype.Azure {
		meta.APIVersion = azure.GetAPIVersion(c)
	}
	if meta.BaseURL == "" {
		meta.BaseURL = channeltype.ChannelBaseURLs[meta.ChannelType]
	}
	meta.APIType = channeltype.ToAPIType(meta.ChannelType)
	return &meta
}
