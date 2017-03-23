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

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	// VersionString is the git describe version set at build time
	VersionString = "?"
	// RevisionString is the git revision set at build time
	RevisionString = "?"
	// GeneratedString is the build date set at build time
	GeneratedString = "?"
)

func init() {
	cli.VersionPrinter = customVersionPrinter
}

func customVersionPrinter(c *cli.Context) {
	fmt.Printf("%v v=%v rev=%v d=%v\n", c.App.Name, VersionString, RevisionString, GeneratedString)
}

func main() {
	app := cli.NewApp()
	app.Usage = "VMware vSphere monitoring utility"
	app.Version = VersionString
	app.Author = "Travis CI GmbH"
	app.Email = "contact+vsphere-monitor@travis-ci.org"

	app.Flags = Flags
	app.Action = mainAction

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func mainAction(c *cli.Context) error {
	log.Print("starting")

	vsphereURL, err := url.Parse(c.String("vsphere-url"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse vSphere URL")
	}

	ctx := context.Background()

	client, err := govmomi.NewClient(ctx, vsphereURL, c.Bool("vsphere-insecure"))
	if err != nil {
		return errors.Wrap(err, "couldn't create govmomi client")
	}

	log.Print("getting hosts")

	finder := find.NewFinder(client.Client, true)

	hosts, err := finder.HostSystemList(ctx, c.String("vsphere-cluster-path"))
	if err != nil {
		return errors.Wrap(err, "couldn't list hosts")
	}

	alarmManager := client.Client.ServiceContent.AlarmManager
	if alarmManager == nil {
		return errors.New("no alarm manager")
	}

	ticker := time.Tick(time.Minute)

	for now := range ticker {
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
		req.SetBasicAuth(c.String("librato-email"), c.String("librato-token"))

		_, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("error sending metrics: %v", err)
		}
	}

	return nil
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
