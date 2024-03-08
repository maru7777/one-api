package util

import (
	"github.com/songquanpeng/one-api/common/config"
	"net/http"
	"time"
)

var HTTPClient *http.Client
var ImpatientHTTPClient *http.Client

func init() {

	tp := &http.Transport{
		IdleConnTimeout: time.Duration(config.IdleTimeout) * time.Second,
	}

	if config.RelayTimeout == 0 {
		HTTPClient = &http.Client{
			Transport: tp,
		}
	} else {
		HTTPClient = &http.Client{
			Transport: tp,
			Timeout:   time.Duration(config.RelayTimeout) * time.Second,
		}
	}

	ImpatientHTTPClient = &http.Client{
		Transport: tp,
		Timeout:   5 * time.Second,
	}
}
