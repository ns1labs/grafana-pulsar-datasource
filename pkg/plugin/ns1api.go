package plugin

import (
	"errors"
	"net/http"
	"time"

	ns1api "gopkg.in/ns1/ns1-go.v2/rest"
)

const (
	timeout = time.Second * 15
	keyAPI  = "apiKey"
)

var (
	errAuthorizationDenied = errors.New("invalid API key")

	httpClient = &http.Client{Timeout: timeout}
)

// checkAPIKey verifies the provided API key against the NS1 API. It returns
// error if the key is invalid, meaning that the authorization was denied.
func checkAPIKey(apiKey string) error {
	var response *http.Response

	client := ns1api.NewClient(httpClient, ns1api.SetAPIKey(apiKey))

	// This will return a 400 error,but we just need to know if the API key
	// is correct.
	_, response, _ = client.PulsarJobs.List("*")
	if response != nil {
		statusCode := response.StatusCode
		if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			return errAuthorizationDenied
		}
	}

	return nil
}
