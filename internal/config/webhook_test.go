package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSettings_Load_Success(t *testing.T) {
	t.Setenv("SYNC_WEBHOOK_SUCCESS_URL", "http://success.example.com")
	t.Setenv("SYNC_WEBHOOK_SUCCESS_METHOD", "POST")
	t.Setenv("SYNC_WEBHOOK_SUCCESS_BODY", "{\"status\":\"ok\"}")
	t.Setenv("SYNC_WEBHOOK_SUCCESS_HEADERS", "Content-Type:application/json,Authorization:Bearer token,X-Custom-Header: CustomValue")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	success := conf.Sync.WebhookSettings.Success
	assert.Equal(t, "http://success.example.com", success.Url)
	assert.Equal(t, "POST", success.Method)
	assert.Equal(t, "{\"status\":\"ok\"}", success.Body)
	assert.Equal(t, map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer token",
		"X-Custom-Header": " CustomValue",
	}, success.Headers)
}

func TestWebhookSettings_Load_Failure(t *testing.T) {
	t.Setenv("SYNC_WEBHOOK_FAILURE_URL", "http://failure.example.com")
	t.Setenv("SYNC_WEBHOOK_FAILURE_METHOD", "PUT")
	t.Setenv("SYNC_WEBHOOK_FAILURE_BODY", "{\"status\":\"error\"}")
	t.Setenv("SYNC_WEBHOOK_FAILURE_HEADERS", "Content-Type:application/json")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	failure := conf.Sync.WebhookSettings.Failure
	assert.Equal(t, "http://failure.example.com", failure.Url)
	assert.Equal(t, "PUT", failure.Method)
	assert.Equal(t, "{\"status\":\"error\"}", failure.Body)
	assert.Equal(t, map[string]string{
		"Content-Type": "application/json",
	}, failure.Headers)
}

func TestWebhookSettings_DefaultValues(t *testing.T) {
	t.Setenv("SYNC_WEBHOOK_SUCCESS_URL", "http://success.example.com")
	t.Setenv("SYNC_WEBHOOK_FAILURE_URL", "http://failure.example.com")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	// Test default values
	assert.Equal(t, "POST", conf.Sync.WebhookSettings.Success.Method)
	assert.Equal(t, "POST", conf.Sync.WebhookSettings.Failure.Method)
	assert.Empty(t, conf.Sync.WebhookSettings.Success.Body)
	assert.Empty(t, conf.Sync.WebhookSettings.Failure.Body)
	assert.Empty(t, conf.Sync.WebhookSettings.Success.Headers)
	assert.Empty(t, conf.Sync.WebhookSettings.Failure.Headers)
}

func TestWebhookSettings_InvalidHeaders(t *testing.T) {
	t.Setenv("SYNC_WEBHOOK_SUCCESS_URL", "http://success.example.com")
	t.Setenv("SYNC_WEBHOOK_SUCCESS_HEADERS", "InvalidHeader")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "process webhook env vars")
}

func TestWebhookSettings_EmptyURLs(t *testing.T) {
	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	assert.Empty(t, conf.Sync.WebhookSettings.Success.Url)
	assert.Empty(t, conf.Sync.WebhookSettings.Failure.Url)
}

func TestWebhookSettings_ClientConfiguration(t *testing.T) {
	t.Setenv("SYNC_WEBHOOK_SUCCESS_URL", "http://success.example.com")
	t.Setenv("CLIENT_TIMEOUT_SECONDS", "30")
	t.Setenv("CLIENT_RETRY_DELAY_SECONDS", "5")
	t.Setenv("CLIENT_SKIP_TLS_VERIFICATION", "true")

	conf := Config{
		Sync: &Sync{},
	}
	err := conf.loadWebhookSettings()
	require.NoError(t, err)

	require.NotNil(t, conf.Sync.WebhookSettings.Client)
	assert.Equal(t, int64(30), conf.Sync.WebhookSettings.Client.Timeout)
	assert.Equal(t, int64(5), conf.Sync.WebhookSettings.Client.RetryDelay)
	assert.True(t, conf.Sync.WebhookSettings.Client.SkipTLSVerification)
}
