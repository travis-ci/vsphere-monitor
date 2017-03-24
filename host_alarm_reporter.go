package vspheremonitor

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var alarmStateToMetricValueMap = map[string]int64{
	"green":  0,
	"yellow": 1,
	"red":    2,
}

// HostAlarmReporter reports the current state of alarms for hosts to Librato
type HostAlarmReporter struct {
	LibratoClient        *LibratoClient
	VSphereClient        *VSphereClient
	AlarmIDMetricNameMap map[string]string
	Clusters             map[string][]*VSphereHost
}

// SetClusterPaths populatest the Clusters attribute of the HostAlarmReporter
// by fetching all the hosts in the clusters with the given inventory paths.
// Note that full inventory paths are required, for example
// "/my-datacenter/host/my-cluster".
func (har *HostAlarmReporter) SetClusterPaths(ctx context.Context, logger logrus.FieldLogger, clusterPaths []string) error {
	logger.Info("getting list of hosts")

	hostCount := 0
	clusters := make(map[string][]*VSphereHost, len(clusterPaths))
	for _, clusterPath := range clusterPaths {
		clusterName := path.Base(clusterPath)

		logger.WithField("cluster_name", clusterName).WithField("cluster_path", clusterPath).Info("getting list of hosts in cluster")
		hosts, err := har.VSphereClient.ListHostsInCluster(ctx, clusterPath)
		if err != nil {
			return errors.Wrapf(err, "error listing hosts in cluster %s (path %s)", clusterName, clusterPath)
		}

		clusters[clusterName] = hosts
		hostCount += len(hosts)
	}
	logger.WithField("host_count", hostCount).Info("found hosts")

	har.Clusters = clusters
	return nil
}

// Report fetches the state of all alarms in the hosts in the Clusters
// attribute, and reports their state to Librato.
func (har *HostAlarmReporter) Report(ctx context.Context, logger logrus.FieldLogger) {
	metrics := har.getMetrics(ctx, logger)
	libratoMetrics := har.convertToLibratoMeasurements(ctx, logger, metrics)

	err := har.LibratoClient.SubmitMeasurements(libratoMetrics)
	if err != nil {
		logger.WithError(err).Error("couldn't submit metrics to Librato")
		return
	}

	logger.WithField("measurement_count", len(libratoMetrics.Gauges)).Info("sent measurements to Librato")
}

func (har *HostAlarmReporter) getMetrics(ctx context.Context, logger logrus.FieldLogger) map[string]map[string]int64 {
	metrics := make(map[string]map[string]int64, len(har.AlarmIDMetricNameMap))
	for _, metricName := range har.AlarmIDMetricNameMap {
		metrics[metricName] = make(map[string]int64)
	}

	for clusterName, hosts := range har.Clusters {
		for _, host := range hosts {
			metricSource := clusterName + "-" + host.Name()

			alarmStates, err := har.VSphereClient.ListAlarmStatesForHost(ctx, host)
			if err != nil {
				logger.WithField("cluster_name", clusterName).WithField("host", host.Name()).WithError(err).Error("error getting alarm states for host")
				continue
			}

			for alarmID, state := range alarmStates {
				metricValue, ok := alarmStateToMetricValueMap[state]
				if ok {
					metrics[alarmID][metricSource] = metricValue
				}
			}
		}
	}

	return metrics
}

func (har *HostAlarmReporter) convertToLibratoMeasurements(ctx context.Context, logger logrus.FieldLogger, metrics map[string]map[string]int64) LibratoMeasurements {
	var libratoMetrics LibratoMeasurements
	libratoMetrics.MeasureTime = time.Now().Unix()

	for alarmID, sourceVals := range metrics {
		for source, value := range sourceVals {
			metricName, ok := har.AlarmIDMetricNameMap[alarmID]
			if !ok {
				continue
			}

			libratoMetrics.Gauges = append(libratoMetrics.Gauges, LibratoGauge{
				Name:   fmt.Sprintf("travis.vsphere-monitor.host-alarm.%s", metricName),
				Value:  float64(value),
				Source: source,
			})
		}
	}

	return libratoMetrics
}
