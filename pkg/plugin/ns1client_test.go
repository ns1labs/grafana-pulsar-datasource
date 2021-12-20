package plugin

import "testing"

var dataPoints = DataPoints{
	Agg: "p50",
	Graph: map[string]DataByASN{
		"US_CA": {
			"123": []DataPoint{
				{1639670400, 35.04938271604939},
				{1639674000, 35.5},
				{1639677600, 34.78125},
				{1639681200, 35},
				{1639684800, 34.5},
				{1639688400, 34.99206349206349},
				{1639692000, 35.72727272727273},
			},
		},
	},
	EndTimestamp:   1639693837,
	StartTimestamp: 1639672237,
	JobID:          "abc",
	AppID:          "xyz",
}

func TestConvertDataPoints(t *testing.T) {
	times, values := ConvertDataPoints("US_CA", "123", dataPoints)

	if len(times) != len(values) {
		t.Errorf("times and values slices are of different lenght")
		return
	}

	data := dataPoints.Graph["US_CA"]["123"]
	for i, dataPoint := range data {
		if times[i].Unix() != int64(dataPoint[0]) {
			t.Errorf("wrong time conversion at index %d", i)
			return
		}
		if values[i] != dataPoint[1] {
			t.Errorf("wrong value conversion at index %d", i)
			return
		}
	}
}
