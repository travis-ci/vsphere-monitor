package vspheremonitor

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

type LibratoClient struct {
	client       *http.Client
	email, token string
}

func NewLibratoClient(email, token string) *LibratoClient {
	return &LibratoClient{
		client: new(http.Client),
		email:  email,
		token:  token,
	}
}

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

type LibratoMeasurements struct {
	MeasureTime int64          `json:"measure_time"`
	Gauges      []LibratoGauge `json:"gauges"`
}

type LibratoGauge struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Source string  `json:"source"`
}
