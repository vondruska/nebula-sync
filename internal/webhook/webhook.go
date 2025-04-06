package webhook

import (
	"fmt"
	"net/http"

	"github.com/lovelaze/nebula-sync/version"
)

type WebhookClient interface {
	Success() error
	Failure() error
}

type webhookClient struct {
	successWebhookURL string
	failureWebhookURL string
}

func NewWebhookClient(successWebhookURL, failureWebhookURL string) WebhookClient {
	return &webhookClient{
		successWebhookURL: successWebhookURL,
		failureWebhookURL: failureWebhookURL,
	}
}

func (webhookClient *webhookClient) Success() error {
	if webhookClient.successWebhookURL == "" {
		return nil
	}
	return invokeWebhook(webhookClient.successWebhookURL)
}

func (webhookClient *webhookClient) Failure() error {
	if webhookClient.failureWebhookURL == "" {
		return nil
	}
	return invokeWebhook(webhookClient.failureWebhookURL)
}

func invokeWebhook(url string) error {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("nebula-sync/%s", version.Version))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
