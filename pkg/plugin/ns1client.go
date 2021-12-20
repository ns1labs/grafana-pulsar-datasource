package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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
		"and and end times don't overlap and the time span in no longer than 30 days")

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

type DataPoint [2]float64
type DataPointSlice []DataPoint
type DataByASN map[string]DataPointSlice

type DataPoints struct {
	Agg            string               `json:"agg"`
	Graph          map[string]DataByASN `json:"graph"`
	EndTimestamp   int64                `json:"end_ts"`
	StartTimestamp int64                `json:"stop_ts"`
	JobID          string               `json:"jobid"`
	AppID          string               `json:"appid"`
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

func (pc *PulsarClient) GetPerformanceData(apiKey string, query queryModel, geo, asn, agg string) ([]time.Time, []float64, error) {
	var (
		parsedAsn int64
		err       error
		resp      *http.Response
	)
	apiClient := pc.getAPIClient(apiKey)

	urlStr := fmt.Sprintf("%spulsar/apps/%s/jobs/%s/data?start=%d&end=%d",
		apiClient.Endpoint.String(),
		query.AppID,
		query.JobID,
		query.From.Unix(),
		query.To.Unix(),
	)

	if len(geo) > 0 && geo != "*" {
		urlStr = fmt.Sprintf("%s&geo=%s", urlStr, geo)
	}

	if asn != "" && asn != "*" {
		parsedAsn, err = strconv.ParseInt(asn, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		urlStr = fmt.Sprintf("%s&asn=%d", urlStr, parsedAsn)
	}

	if len(agg) > 0 {
		urlStr = fmt.Sprintf("%s&agg=%s", urlStr, agg)
	}

	apiURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}

	req := &http.Request{
		Method: http.MethodGet,
		URL:    apiURL,
		Header: map[string][]string{
			"X-NSONE-Key": []string{apiKey},
		},
	}

	resp, err = httpClient.Do(req)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var dataPoints DataPoints
	err = json.Unmarshal(body, &dataPoints)
	if err != nil {
		return nil, nil, err
	}
	times, values := ConvertDataPoints(geo, asn, dataPoints)
	// json := fmt.Sprintf("%s", string(body))
	// Logger.Info(json)

	return times, values, nil
}

func ConvertDataPoints(geo, asn string, dp DataPoints) ([]time.Time, []float64) {
	var (
		times  []time.Time
		values []float64
	)

	data := dp.Graph[geo][asn]
	times = make([]time.Time, len(data))
	values = make([]float64, len(data))

	for i, dataPoint := range data {
		times[i] = time.Unix(int64(dataPoint[0]), 0)
		values[i] = dataPoint[1]
	}

	return times, values
}

// NewPulsarClient is the default constructor for the Pulsar Client object.
func NewPulsarClient() *PulsarClient {
	return &PulsarClient{
		apiClientCache: make(map[string]*ns1api.Client),
	}
}
