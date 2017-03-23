package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	vspheremonitor "github.com/travis-ci/vsphere-monitor"
	"github.com/urfave/cli"
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
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	logger := logrus.WithField("pid", os.Getpid())

	logger.WithField("version", VersionString).Info("starting vsphere-monitor")
	defer logger.Info("stopping vsphere-monitor")

	vsphereURL, err := url.Parse(c.String("vsphere-url"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse vSphere URL")
	}

	ctx := context.Background()

	vSphereClient, err := vspheremonitor.NewVSphereClient(ctx, vsphereURL, c.Bool("vsphere-insecure"))
	if err != nil {
		return errors.Wrap(err, "error creating vsphere client")
	}

	logger.Info("getting list of hosts")

	hostCount := 0
	clusters := make(map[string][]*vspheremonitor.VSphereHost, len(c.StringSlice("vsphere-cluster-path")))
	for _, clusterPath := range c.StringSlice("vsphere-cluster-path") {
		clusterName := path.Base(clusterPath)

		logger.WithField("cluster_name", clusterName).WithField("cluster_path", clusterPath).Info("getting list of hosts in cluster")
		hosts, err := vSphereClient.ListHostsInCluster(ctx, clusterPath)
		if err != nil {
			return errors.Wrapf(err, "error listing hosts in cluster %s (path %s)", clusterName, clusterPath)
		}

		clusters[clusterName] = hosts
		hostCount += len(hosts)
	}
	logger.WithField("host_count", hostCount).Info("found hosts")

	libratoClient := vspheremonitor.NewLibratoClient(c.String("librato-email"), c.String("librato-token"))

	alertIdMetricNameMap := make(map[string]string, len(c.StringSlice("vsphere-host-alert-id-metric-name")))
	for _, alertIdMetricName := range c.StringSlice("vsphere-host-alert-id-metric-name") {
		parts := strings.SplitN(alertIdMetricName, ":", 2)
		alertIdMetricNameMap[parts[0]] = parts[1]
	}

	ticker := time.Tick(time.Minute)

	for now := range ticker {
		metrics := make(map[string]map[string]int64, len(alertIdMetricNameMap))
		for _, metricName := range alertIdMetricNameMap {
			metrics[metricName] = make(map[string]int64)
		}

		for clusterName, hosts := range clusters {
			for _, host := range hosts {
				metricSource := clusterName + "-" + host.Name()

				alarmStates, err := vSphereClient.ListAlarmStatesForHost(ctx, host)
				if err != nil {
					logger.WithField("cluster_name", clusterName).WithField("host", host.Name()).WithError(err).Error("error getting alarm states for host")
					continue
				}

				for alarmID, state := range alarmStates {
					metricName, ok := alertIdMetricNameMap[alarmID]
					if !ok {
						continue
					}

					switch state {
					case "green":
						metrics[metricName][metricSource] = 0
					case "yellow":
						metrics[metricName][metricSource] = 1
					case "red":
						metrics[metricName][metricSource] = 2
					case "gray":
						// no data, so do nothing
					}
				}
			}
		}

		var libratoMetrics vspheremonitor.LibratoMeasurements
		libratoMetrics.MeasureTime = now.Unix()

		measurementCount := 0
		for name, sourceVals := range metrics {
			if len(sourceVals) == 0 {
				continue
			}

			for source, value := range sourceVals {
				measurementCount++
				libratoMetrics.Gauges = append(libratoMetrics.Gauges, vspheremonitor.LibratoGauge{
					Name:   fmt.Sprintf("travis.vsphere-monitor.host-alarm.%s", name),
					Value:  float64(value),
					Source: source,
				})
			}
		}

		err := libratoClient.SubmitMeasurements(libratoMetrics)
		if err != nil {
			logger.WithError(err).Error("couldn't submit metrics to Librato")
		}

		logger.WithField("measurement_count", measurementCount).Info("sent measurements to Librato")
	}

	return nil
}
