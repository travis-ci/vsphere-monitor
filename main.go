package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

func main() {
	log.Print("starting")

	vsphereURL, err := url.Parse(os.Getenv("VSPHERE_URL"))
	if err != nil {
		log.Fatalf("couldn't parse vSphere URL: %v", err)
	}

	ctx := context.Background()

	client, err := govmomi.NewClient(ctx, vsphereURL, true)
	if err != nil {
		log.Fatalf("couldn't create govmomi client: %v", err)
	}

	log.Print("getting hosts")

	finder := find.NewFinder(client.Client, true)

	hosts, err := finder.HostSystemList(ctx, os.Getenv("VSPHERE_CLUSTER_PATH"))
	if err != nil {
		log.Fatalf("couldn't list hosts: %v", err)
	}

	alarmManager := client.Client.ServiceContent.AlarmManager
	if alarmManager == nil {
		log.Fatal("no alarm manager")
	}

	c := time.Tick(time.Minute)

	for now := range c {
		metrics := make(map[string]map[string]int64)

		for _, host := range hosts {
			alarmStateResp, err := methods.GetAlarmState(ctx, client.Client, &types.GetAlarmState{This: *alarmManager, Entity: host.Reference()})
			if err != nil {
				fmt.Printf("couldn't get alarm states for host %s: %v\n", host.Name(), err)
				continue
			}

			for _, state := range alarmStateResp.Returnval {
				if _, ok := metrics[state.Alarm.Value]; !ok {
					metrics[state.Alarm.Value] = make(map[string]int64)
				}

				switch state.OverallStatus {
				case types.ManagedEntityStatusGreen:
					metrics[state.Alarm.Value][host.Name()] = 0
				case types.ManagedEntityStatusYellow:
					metrics[state.Alarm.Value][host.Name()] = 1
				case types.ManagedEntityStatusRed:
					metrics[state.Alarm.Value][host.Name()] = 2
				case types.ManagedEntityStatusGray:
				}
			}
		}

		var libratoMetrics libratoMeasurements
		libratoMetrics.MeasureTime = now.Unix()

		for name, sourceVals := range metrics {
			if len(sourceVals) == 0 {
				continue
			}

			for source, value := range sourceVals {
				libratoMetrics.Gauges = append(libratoMetrics.Gauges, libratoGauge{
					Name:   fmt.Sprintf("travis.vsphere-monitor.host-alarm.%s", name),
					Value:  float64(value),
					Source: source,
				})
			}
		}

		body, err := json.Marshal(libratoMetrics)
		if err != nil {
			log.Printf("couldn't marshal metrics: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", "https://metrics-api.librato.com/v1/metrics", bytes.NewReader(body))
		if err != nil {
			log.Printf("couldn't create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(os.Getenv("LIBRATO_EMAIL"), os.Getenv("LIBRATO_TOKEN"))

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("error sending metrics: %v", err)
		}
	}
}

type libratoMeasurements struct {
	MeasureTime int64          `json:"measure_time"`
	Gauges      []libratoGauge `json:"gauges"`
}

type libratoGauge struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Source string  `json:"source"`
}
