package webhook

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lovelaze/nebula-sync/internal/config"
	"github.com/lovelaze/nebula-sync/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookClient(t *testing.T) {
	t.Run("success webhook uses success configuration", func(t *testing.T) {
		// Setup test server to verify request
		var receivedHeaders http.Header
		var receivedBody string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			buf := make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			receivedBody = string(buf[:n])
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// Create webhook settings
		settings := &config.WebhookSettings{
			Success: config.WebhookEventSetting{
				Url:     ts.URL,
				Method:  "POST",
				Body:    "success-body",
				Headers: map[string]string{"X-Test": "success"},
			},
			Client: config.WebhookClient{},
		}

		client := NewWebhookClient(settings)
		err := client.Success()
		require.NoError(t, err)

		// Verify request
		assert.Equal(t, "success-body", receivedBody)
		assert.Equal(t, "success", receivedHeaders.Get("X-Test"))
		assert.Equal(t, "nebula-sync/"+version.Version, receivedHeaders.Get("User-Agent"))
	})

	t.Run("failure webhook uses failure configuration", func(t *testing.T) {
		var receivedHeaders http.Header
		var receivedBody string
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			buf := make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			receivedBody = string(buf[:n])
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		settings := &config.WebhookSettings{
			Failure: config.WebhookEventSetting{
				Url:     ts.URL,
				Method:  "PUT",
				Body:    "failure-body",
				Headers: map[string]string{"X-Test": "failure"},
			},
			Client: config.WebhookClient{},
		}

		client := NewWebhookClient(settings)
		err := client.Failure()
		require.NoError(t, err)

		assert.Equal(t, "failure-body", receivedBody)
		assert.Equal(t, "failure", receivedHeaders.Get("X-Test"))
	})

	t.Run("empty url skips webhook", func(t *testing.T) {
		settings := &config.WebhookSettings{
			Success: config.WebhookEventSetting{
				Url: "",
			},
		}

		client := NewWebhookClient(settings)
		err := client.Success()
		require.NoError(t, err)
	})

	t.Run("error on non-200 response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		settings := &config.WebhookSettings{
			Success: config.WebhookEventSetting{
				Url: ts.URL,
			},
		}

		client := NewWebhookClient(settings)
		err := client.Success()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "webhook returned status 400")
	})
}
