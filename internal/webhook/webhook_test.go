package webhook

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lovelaze/nebula-sync/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WebhookTestSuite struct {
	suite.Suite
	server *httptest.Server
}

func TestWebhookSuite(t *testing.T) {
	suite.Run(t, new(WebhookTestSuite))
}

func (suite *WebhookTestSuite) SetupTest() {
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedUA := fmt.Sprintf("nebula-sync/%s", version.Version)
		if r.Header.Get("User-Agent") != expectedUA {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func (suite *WebhookTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *WebhookTestSuite) TestEmptyURLs() {
	client := NewWebhookClient("", "")

	err := client.Success()
	assert.NoError(suite.T(), err)

	err = client.Failure()
	assert.NoError(suite.T(), err)
}

func (suite *WebhookTestSuite) TestSuccessWebhook() {
	client := NewWebhookClient(suite.server.URL, "")

	err := client.Success()
	assert.NoError(suite.T(), err)
}

func (suite *WebhookTestSuite) TestFailureWebhook() {
	client := NewWebhookClient("", suite.server.URL)

	err := client.Failure()
	assert.NoError(suite.T(), err)
}

func (suite *WebhookTestSuite) TestInvalidURL() {
	client := NewWebhookClient("invalid-url", "")

	err := client.Success()
	assert.Error(suite.T(), err)
}

func (suite *WebhookTestSuite) TestServerError() {
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	client := NewWebhookClient(errorServer.URL, "")

	err := client.Success()
	assert.Error(suite.T(), err)
}

func (suite *WebhookTestSuite) TestServerUnavailable() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client := NewWebhookClient(server.URL, "")

	err := client.Success()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "send webhook request")
}

func (suite *WebhookTestSuite) TestUserAgentHeader() {
	var receivedUA string
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	client := NewWebhookClient(testServer.URL, "")
	err := client.Success()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf("nebula-sync/%s", version.Version), receivedUA)
}
