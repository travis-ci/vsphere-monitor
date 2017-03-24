package vspheremonitor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// A LibratoClient allows for submitting measurements to Librato using the
// Librato API. It only supports accounts using the "legacy" source-based
// metrics, tag-based metrics are not supported.
type LibratoClient struct {
	client       *http.Client
	email, token string
}

// NewLibratoClient creates a new LibratoClient using the given email and
// token. The token needs to be associated with the Librato account with the
// given email and needs to have record permissions.
func NewLibratoClient(email, token string) *LibratoClient {
	return &LibratoClient{
		client: new(http.Client),
		email:  email,
		token:  token,
	}
}

// SubmitMeasurements submits a list of measurements to Librato. All fields in
// the given structure are required.
func (lc *LibratoClient) SubmitMeasurements(measurements LibratoMeasurements) error {
	body, err := json.Marshal(measurements)
	if err != nil {
		return errors.Wrap(err, "error marshalling measurements to JSON")
	}

	req, err := http.NewRequest("POST", "https://metrics-api.librato.com/v1/metrics", bytes.NewReader(body))
	if err != nil {
		return errors.Wrap(err, "error creating HTTP request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(lc.email, lc.token)

	_, err = lc.client.Do(req)
	return errors.Wrap(err, "error while sending HTTP request")
}

// LibratoMeasurements contains a set of metric measurements taken at a given
// timestamp, to be submitted to Librato using
// LibratoClient.SubmitMeasurements.
type LibratoMeasurements struct {
	// MeasureTime is the unix timestamp for when the measurements were taken
	MeasureTime int64 `json:"measure_time"`

	// Gauges is a list of gauge measurements.
	Gauges []LibratoGauge `json:"gauges"`
}

// LibratoGauge contains an individual gauge measurement.
type LibratoGauge struct {
	// Name is the name of the metric the measurement is associated with. The
	// name must be 255 or fewer characters, and may only consist of
	// 'A-Za-z0-9.:-_'.
	Name string `json:"name"`

	// Value is the current value of the measurement, as of the time specified
	// in MeasureTime in the LibratoMeasurement structure.
	Value float64 `json:"value"`

	// Source is a name describing the originating source of the measurement.
	// The source name must be 255 or fewer characters, and may only consist of
	// 'A-Za-z0-9.:-_'.
	Source string `json:"source"`
}
