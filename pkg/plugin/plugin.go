package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var Logger = log.DefaultLogger

// Make sure PulsarDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*PulsarDatasource)(nil)
	_ backend.CheckHealthHandler    = (*PulsarDatasource)(nil)
	_ backend.StreamHandler         = (*PulsarDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*PulsarDatasource)(nil)

	errDataSourceInstanceSettingsNil = errors.New("data source instance settings not present in the plugin context")
	errDecryptedSecureDataNil        = errors.New("secure decrypted data not found")
	errAPIKeyNotFound                = errors.New("NS1 API key not found")
)

type queryModel struct {
	AppID      string `json:"appid"`
	JobID      string `json:"jobid"`
	MetricType string `json:"metricType"`
	From,
	To time.Time
}

// NewPulsarDatasource creates a new datasource instance.
func NewPulsarDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &PulsarDatasource{}, nil
}

// PulsarDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type PulsarDatasource struct {
	pulsarClient *PulsarClient
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewPulsarDatasource factory function.
func (p *PulsarDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (p *PulsarDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	if p.pulsarClient == nil {
		p.pulsarClient = NewPulsarClient()
	}

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := p.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func buildLabel(appName, appID, jobID, metric, geo, asn string) string {
	return fmt.Sprintf("%s (%s):%s:%s:%s:%s", appName, appID, jobID, metric, geo, asn)
}

func (p *PulsarDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var (
		qm       queryModel
		response backend.DataResponse
		geo      = "*"
		asn      = "*"
		agg      = "p99"
		times    = []time.Time{query.TimeRange.From, query.TimeRange.To}
		values   = []float64{10, 20}
	)

	apiKey, err := getAPIKeyFromContext(pCtx)
	if err != nil {
		response.Error = err
		return response
	}

	// Unmarshal the JSON into our queryModel.
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// create data frame response.
	frame := data.NewFrame("response")

	// test
	geo = "US"
	qm.From = query.TimeRange.From
	qm.To = query.TimeRange.To

	if qm.AppID != "" && qm.JobID != "" {
		times, values, err = p.pulsarClient.GetPerformanceData(apiKey, qm, geo, asn, agg)
	}

	// add fields.
	dataLabel := buildLabel("Akamai", qm.AppID, qm.JobID, "Job Name", geo, asn)
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
		data.NewField(dataLabel, nil, values),
	)

	apps, err := p.pulsarClient.GetApps(apiKey, OptionAppFetchJobs(true))
	if err != nil {
		response.Error = err
		return response
	}

	frame.Meta = &data.FrameMeta{Custom: apps}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (p *PulsarDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var (
		apiKey string
		err    error
		client *PulsarClient
	)

	if req == nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "backend received a nil request",
		}, nil
	}

	apiKey, err = getAPIKeyFromContext(req.PluginContext)
	if err != nil {
		if errors.Is(err, errDataSourceInstanceSettingsNil) {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "backend received nil settings",
			}, nil
		}
		if errors.Is(err, errDecryptedSecureDataNil) {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: err.Error(),
			}, nil
		}
		if errors.Is(err, errAPIKeyNotFound) {
			return &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "API key not present",
			}, nil
		}
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	client = NewPulsarClient()

	if err := client.CheckAPIKey(apiKey); err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}, nil
	}

	if p.pulsarClient == nil {
		p.pulsarClient = client
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source status correct",
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (p *PulsarDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	if req.Path == "stream" {
		// Allow subscribing only on expected path.
		status = backend.SubscribeStreamStatusOK
	}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (p *PulsarDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")

	// Add fields (matching the same schema used in QueryData).
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField("values", nil, make([]int64, 1)),
	)

	counter := 0

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case <-time.After(time.Second):
			// Send new data periodically.
			frame.Fields[0].Set(0, time.Now())
			frame.Fields[1].Set(0, int64(10*(counter%2+1)))

			counter++

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (p *PulsarDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func getAPIKeyFromContext(pluginContext backend.PluginContext) (string, error) {
	if pluginContext.DataSourceInstanceSettings == nil {
		return "", errDataSourceInstanceSettingsNil
	}

	dsis := pluginContext.DataSourceInstanceSettings
	if dsis.DecryptedSecureJSONData == nil {
		return "", errDecryptedSecureDataNil
	}

	apiKey, exists := dsis.DecryptedSecureJSONData[APIKey]
	if !exists {
		return "", errAPIKeyNotFound
	}

	return apiKey, nil
}
