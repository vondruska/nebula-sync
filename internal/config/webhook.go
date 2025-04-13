package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type WebhookSettings struct {
	Failure WebhookEventSetting `ignored:"true"`
	Success WebhookEventSetting `ignored:"true"`
	Client  Client              `ignored:"true"`
}

type WebhookEventSetting struct {
	Body    string            `default:"" envconfig:"BODY"`
	Headers map[string]string `default:"" envconfig:"HEADERS"`
	Method  string            `default:"POST" envconfig:"METHOD"`
	Url     string            `default:"" envconfig:"URL"`
}

const envPrefix = "SYNC_WEBHOOK_"

func (c *Config) loadWebhookSettings() error {
	webhookSettings := WebhookSettings{
		Failure: WebhookEventSetting{},
		Success: WebhookEventSetting{},
		Client:  Client{},
	}

	if err := envconfig.Process(envPrefix+"FAILURE", &webhookSettings.Failure); err != nil {
		return fmt.Errorf("process webhook env vars for failure: %w", err)
	}
	if err := envconfig.Process(envPrefix+"SUCCESS", &webhookSettings.Success); err != nil {
		return fmt.Errorf("process webhook env vars for success: %w", err)
	}

	if err := envconfig.Process(envPrefix+"CLIENT", &webhookSettings.Client); err != nil {
		return fmt.Errorf("process webhook env vars for client: %w", err)
	}

	c.Sync.WebhookSettings = &webhookSettings

	return nil
}
