package types

import (
	"errors"
	"net/url"
	"strings"
)

type WebhookConfig struct {
	URL    string `json:"url" yaml:"url"`
	Method string `json:"method" yaml:"method"`
}

func (wc *WebhookConfig) Validate() error {
	parsedUrl, err := url.ParseRequestURI(wc.URL)
	if err != nil {
		return errors.New("webhook url format invalid")
	}

	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return errors.New("webhook scheme invalid")
	}

	if parsedUrl.Host == "" {
		return errors.New("webhook host invalid")
	}

	if !strings.EqualFold(wc.Method, "POST") {
		return errors.New("unsupported http method for webhook")
	}

	return nil
}
