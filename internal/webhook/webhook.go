package webhook

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/rs/zerolog/log"
)

type WebhookClient interface {
	Success() error
	Failure() error
}

type webhookClient struct {
	successConfig config.WebhookEventSetting
	failureConfig config.WebhookEventSetting
	client        *http.Client
}

func NewWebhookClient(c *config.WebhookSettings) WebhookClient {
	return &webhookClient{
		successConfig: c.Success,
		failureConfig: c.Failure,
		client:        c.Client.NewHttpClient(),
	}
}

func (webhookClient *webhookClient) Success() error {
	return invokeWebhook(webhookClient.client, webhookClient.successConfig)
}

func (webhookClient *webhookClient) Failure() error {
	return invokeWebhook(webhookClient.client, webhookClient.failureConfig)
}

func invokeWebhook(client *http.Client, settings config.WebhookEventSetting) error {
	if settings.Url == "" {
		return nil
	}

	log.Debug().
		Str("url", settings.Url).
		Str("method", settings.Method).
		Str("body", settings.Body).
		Interface("headers", settings.Headers).
		Msg("Invoking webhook")

	req, err := http.NewRequest(settings.Method, settings.Url, strings.NewReader(settings.Body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("nebula-sync/%s", version.Version))

	for key, value := range settings.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
