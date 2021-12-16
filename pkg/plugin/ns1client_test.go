//go:build integration
// +build integration

package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func getApiKey(t *testing.T) string {
	apiKey, found := os.LookupEnv("NS1_API_KEY")
	if !found {
		t.Fatal("API KEY not found")
	}
	return apiKey
}

func TestPulsarClient_GetPulsarApps(t *testing.T) {
	apiKey := getApiKey(t)
	client := NewPulsarClient(apiKey)

	apps, err := client.GetApps(PulsarAppFetchJobs(true))
	if err != nil {
		t.Errorf("error getting pulsar apps: %v", err)
		return
	}

	var appsBytes []byte
	appsBytes, err = json.MarshalIndent(apps, "", "  ")
	if err != nil {
		t.Errorf("error marshalling apps to json: %v", err)
		return
	}

	fmt.Printf("%s\n", string(appsBytes))
}
