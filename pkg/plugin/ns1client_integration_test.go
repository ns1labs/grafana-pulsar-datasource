//go:build integration
// +build integration

package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var (
	query = backend.DataQuery{
		RefID:         "A",
		QueryType:     "",
		MaxDataPoints: 1246,
		Interval:      15000000000,
		TimeRange: backend.TimeRange{
			From: time.Unix(1640001600, 0),
			To:   time.Unix(1640023200, 0),
		},
		JSON: []byte(`
			"datasource": {
				"type": "ns1labs-pulsar-datasource",
				"uid":  "mh-5FKhnz",
			},
			"datasourceId":  4,
			"intervalMs":    15000,
			"maxDataPoints": 1246,
			"refId":         "A",
			}`),
	}
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
	client := NewPulsarClient()

	apps, err := client.GetApps(apiKey, OptionAppFetchJobs(true))
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
