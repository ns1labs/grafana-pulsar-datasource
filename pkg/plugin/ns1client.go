package plugin

import (
	"errors"
	"net/http"
	"sync"
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
	apiClientCache map[string]*ns1api.Client
	apiClientLock  sync.RWMutex
	data           *PulsarData
}

// getAPIClient maintains a local cache of the NS1 api clients for each API key
// handled. This way we can set the api key at the QueryEditor level.
func (pc *PulsarClient) getAPIClient(apiKey string) *ns1api.Client {
	pc.apiClientLock.Lock()
	defer pc.apiClientLock.Unlock()

	client, exists := pc.apiClientCache[apiKey]
	if !exists {
		client = ns1api.NewClient(httpClient, ns1api.SetAPIKey(apiKey))
		pc.apiClientCache[apiKey] = client
	}

	return client
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
	pc.apiClientLock.Lock()
	pc.apiClientCache[apiKey] = client
	pc.apiClientLock.Unlock()

	return nil
}

func OptionAppFetchJobs(fetchJobs bool) PulsarAppParameter {
	return func(p *PulsarAppParameters) {
		p.FetchJobs = fetchJobs
	}
}

func PulsarAppFetchInactive(fetchInactive bool) PulsarAppParameter {
	return func(p *PulsarAppParameters) {
		p.FetchInactiveApps = fetchInactive
	}
}

func (pc *PulsarClient) GetApps(apiKey string, params ...PulsarAppParameter) ([]App, error) {
	var (
		pulsarApps []*pulsar.Application
		err        error
		apps       []App
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

	apiClient := pc.getAPIClient(apiKey)

	pulsarApps, _, err = apiClient.Applications.List()
	if err != nil {
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
			apps[i].Jobs, err = pc.GetJobs(apiKey, pulsarApp.ID, params...)
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

// OptionJobsFetchInactive indicates that the API must also retrieve jobs
// marked as inactive along with active ones.
func OptionJobsFetchInactive(fetchInactive bool) PulsarAppParameter {
	return func(p *PulsarAppParameters) {
		p.FetchInactiveJobs = fetchInactive
	}
}

func (pc *PulsarClient) GetJobs(apiKey, appID string, params ...PulsarAppParameter) ([]Job, error) {
	var (
		jobs  []Job
		err   error
		pjobs []*pulsar.PulsarJob
	)

	apiClient := pc.getAPIClient(apiKey)
	pjobs, _, err = apiClient.PulsarJobs.List(appID)
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
func NewPulsarClient() *PulsarClient {
	return &PulsarClient{
		apiClientCache: make(map[string]*ns1api.Client),
	}
}
