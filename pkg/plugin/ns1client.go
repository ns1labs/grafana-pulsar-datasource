package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	ns1api "gopkg.in/ns1/ns1-go.v2/rest"
	"gopkg.in/ns1/ns1-go.v2/rest/model/pulsar"
)

const (
	timeout                = time.Second * 15
	APIKey                 = "apiKey"
	metricTypePerformance  = "performance"
	metricTypeAvailability = "availability"
	metricTypeDecisions    = "decisions"
)

var (
	errAuthorizationDenied = errors.New("invalid API key")
	errDataRetrieval       = errors.New("error retrieving data, make sure start " +
		"and and end times don't overlap and the time span it's no longer than 30 days")
	errNoDataFound = errors.New("no data found")

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

// GetAppsResponse holds the Apps and Jobs info in two formats: A slice for the
// UI and a couple of maps for internal use.
type GetAppsResponse struct {
	Apps    []App
	AppsMap map[string]App
	JobsMap map[string]Job
}

type PulsarAppParameters struct {
	FetchInactiveApps bool
	FetchJobs         bool
	FetchInactiveJobs bool
}

type PulsarAppParameter func(p *PulsarAppParameters)

// PulsarData is the data struct for passing apps and jobs info to the Frontend.
type PulsarData struct {
	Applications *GetAppsResponse `json:"applications,omitempty"`
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
		client = ns1api.NewClient(
			&http.Client{Timeout: timeout},
			ns1api.SetAPIKey(apiKey),
		)
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

func (pc *PulsarClient) GetApps(apiKey string, params ...PulsarAppParameter) (*GetAppsResponse, error) {
	var (
		pulsarApps []*pulsar.Application
		err        error
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

	appsResponse := &GetAppsResponse{
		Apps:    make([]App, len(pulsarApps)),
		AppsMap: make(map[string]App),
		JobsMap: make(map[string]Job),
	}

	for i, pulsarApp := range pulsarApps {
		if !pulsarApp.Active && !parameters.FetchInactiveApps {
			// skip inactive apps
			continue
		}
		appsResponse.Apps[i] = App{
			AppID: pulsarApp.ID,
			Name:  pulsarApp.Name,
			Jobs:  []Job{},
		}
		appsResponse.AppsMap[pulsarApp.ID] = appsResponse.Apps[i]

		if parameters.FetchJobs {
			appsResponse.Apps[i].Jobs, err = pc.GetJobs(apiKey, pulsarApp.ID, params...)
			if err != nil {
				return nil, err
			}
			for _, j := range appsResponse.Apps[i].Jobs {
				appsResponse.JobsMap[j.JobID] = j
			}
		}
	}

	// replace current data
	pc.data = &PulsarData{
		Applications: appsResponse,
		// Calculate new TTLs
	}

	return appsResponse, nil
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

func (pc *PulsarClient) buildURL(endpoint string, qm *queryModel) (*url.URL, error) {
	var urlStr string

	if qm.MetricType == metricTypePerformance {
		urlStr = fmt.Sprintf("%spulsar/query/performance/time", endpoint)
	} else {
		urlStr = fmt.Sprintf("%spulsar/query/availability/time", endpoint)
	}

	urlStr = fmt.Sprintf("%s?start=%d&end=%d&jobs=%s", urlStr,
		qm.From.Unix(), qm.To.Unix(), qm.JobID)

	if len(qm.Aggregation) > 0 {
		urlStr = fmt.Sprintf("%s&agg=%s", urlStr, qm.Aggregation)
	}
	if qm.Geo == "*" {
		urlStr = fmt.Sprintf("%s&area=GLOBAL", urlStr)
	} else {
		urlStr = fmt.Sprintf("%s&area=%s", urlStr, qm.Geo)
	}
	if qm.ASN != "*" {
		urlStr = fmt.Sprintf("%s&asn=%s", urlStr, qm.ASN)
	}

	return url.Parse(urlStr)
}

func (pc *PulsarClient) GetData(apiKey string, query *queryModel) ([]time.Time, []float64, error) {
	var (
		apiURL *url.URL
		resp   *http.Response
		err    error
		times  []time.Time
		values []float64
		body   []byte
		offset int64
	)

	apiClient := pc.getAPIClient(apiKey)

	if apiURL, err = pc.buildURL(apiClient.Endpoint.String(), query); err != nil {
		return nil, nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    apiURL,
		Header: map[string][]string{
			"X-NSONE-Key": []string{apiKey},
		},
	}

	if resp, err = httpClient.Do(req); err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	// This error can be returned by the API.
	if resp.StatusCode == http.StatusBadRequest {
		return nil, nil, errDataRetrieval
	}

	if body, err = io.ReadAll(resp.Body); err != nil {
		return nil, nil, err
	}

	data := make([]map[string]float64, 0)
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, nil, err
	}

	size := int64(len(data))
	if size == 0 {
		return nil, nil, errNoDataFound
	}
	totalSize := size

	if query.MaxDataPoints < size {
		offset = size - query.MaxDataPoints
		size = query.MaxDataPoints
	}

	times = make([]time.Time, size)
	values = make([]float64, size)
	var idx int

	// Retrieve the latest data
	for i := offset; i < totalSize; i++ {
		dataPoint := data[i]
		times[idx] = time.Unix(int64(dataPoint["timestamp"]), 0)
		values[idx] = dataPoint[query.JobID]
		idx++
	}

	return times, values, nil
}

// NewPulsarClient is the default constructor for the Pulsar Client object.
func NewPulsarClient() *PulsarClient {
	return &PulsarClient{
		apiClientCache: make(map[string]*ns1api.Client),
	}
}
