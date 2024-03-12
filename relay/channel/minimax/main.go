package minimax

import (
	"fmt"

	"github.com/Laisky/errors/v2"
	"github.com/songquanpeng/one-api/relay/constant"
	"github.com/songquanpeng/one-api/relay/util"
)

func GetRequestURL(meta *util.RelayMeta) (string, error) {
	if meta.Mode == constant.RelayModeChatCompletions {
		return fmt.Sprintf("%s/v1/text/chatcompletion_v2", meta.BaseURL), nil
	}
	return "", errors.Errorf("unsupported relay mode %d for minimax", meta.Mode)
}
