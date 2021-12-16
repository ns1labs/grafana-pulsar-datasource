package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	ns1api "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/pulsar"
)

const (
	timeout = time.Second * 15
	APIKey  = "apiKey"
)

var (
	errAuthorizationDenied = errors.New("invalid API key")

	httpClient = &http.Client{Timeout: timeout}
)

type Job struct {
	JobID string `json:"jobid"`
	Name  string `json:"name"`
}

// App is a basic model to exchange information with the frontend.
type App struct {
	AppID string `json:"appid"`
	Name  string `json:"name,omitempty"`
	Jobs  []Job  `json:"jobs"`
}

type PulsarAppParameters struct {
	FetchInactiveApps bool
	FetchJobs         bool
	FetchInactiveJobs bool
}

type PulsarAppParameter func(p *PulsarAppParameters)

// PulsarData is the data struct for passing apps and jobs info to the Frontend.
type PulsarData struct {
	Applications []App `json:"applications,omitempty"`
	ttl          time.Duration
	expiresOn    time.Time
}

type PulsarClient struct {
	apiKey    string
	apiClient *ns1api.Client
	data      *PulsarData
}

// CheckAPIKey verifies the provided API key against the NS1 API. It returns
// error if the key is invalid, meaning that the authorization was denied.
func (pc *PulsarClient) CheckAPIKey(apiKey string) error {
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

	// Update the client as the api key has changed
	pc.apiClient = client

	return nil
}

func PulsarAppFetchJobs(fetchJobs bool) PulsarAppParameter {
	return func(p *PulsarAppParameters) {
		p.FetchJobs = fetchJobs
	}
}

func PulsarAppFetchInactive(fetchInactive bool) PulsarAppParameter {
	return func(p *PulsarAppParameters) {
		p.FetchInactiveApps = fetchInactive
	}
}

func (pc *PulsarClient) GetApps(params ...PulsarAppParameter) ([]App, error) {
	var (
		req    *http.Request
		resp   *http.Response
		ns1Url *url.URL
		body   []byte
		err    error
		apps   []App
	)

	if pc.data != nil {
		// Verify the TTL to refresh
		return pc.data.Applications, nil
	}

	parameters := &PulsarAppParameters{
		FetchInactiveApps: false,
		FetchJobs:         false,
	}
	for _, param := range params {
		param(parameters)
	}

	ns1Url, err = url.Parse(
		fmt.Sprintf("%spulsar/apps", pc.apiClient.Endpoint.String()),
	)
	if err != nil {
		return nil, err
	}

	req = &http.Request{
		Method: http.MethodGet,
		URL:    ns1Url,
		Header: map[string][]string{
			"X-NSONE-Key": []string{pc.apiKey},
		},
	}

	if resp, err = httpClient.Do(req); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if body, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	pulsarApps := make([]pulsar.Application, 0)
	if err = json.Unmarshal(body, &pulsarApps); err != nil {
		return nil, err
	}

	apps = make([]App, len(pulsarApps))
	for i, pulsarApp := range pulsarApps {
		if parameters.FetchInactiveApps {
			// skip inactive apps
			continue
		}
		apps[i] = App{
			AppID: pulsarApp.ID,
			Name:  pulsarApp.Name,
			Jobs:  []Job{},
		}

		if parameters.FetchJobs {
			apps[i].Jobs, err = pc.GetJobs(pulsarApp.ID, params...)
			if err != nil {
				return nil, err
			}
		}
	}

	// replace current data
	pc.data = &PulsarData{
		Applications: apps,
	}

	return apps, nil
}

func PulsarJobsFetchInactive(fetchInactive bool) PulsarAppParameter {
	return func(p *PulsarAppParameters) {
		p.FetchInactiveJobs = fetchInactive
	}
}

func (pc *PulsarClient) GetJobs(appID string, params ...PulsarAppParameter) ([]Job, error) {
	var (
		jobs  []Job
		err   error
		pjobs []*pulsar.PulsarJob
	)

	pjobs, _, err = pc.apiClient.PulsarJobs.List(appID)
	if err != nil {
		return nil, err
	}

	parameters := PulsarAppParameters{}
	for _, param := range params {
		param(&parameters)
	}

	jobs = make([]Job, len(pjobs))
	for i, pjob := range pjobs {
		if parameters.FetchInactiveJobs {
			continue
		}

		jobs[i] = Job{
			JobID: pjob.JobID,
			Name:  pjob.Name,
		}
	}

	return jobs, nil
}

// NewPulsarClient is the default constructor for the Pulsar Client object.
func NewPulsarClient(apiKey string) *PulsarClient {
	apiClient := ns1api.NewClient(httpClient, ns1api.SetAPIKey(apiKey))

	return &PulsarClient{
		apiKey:    apiKey,
		apiClient: apiClient,
	}
}
