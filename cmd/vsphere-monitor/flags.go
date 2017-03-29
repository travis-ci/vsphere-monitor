package main

import "github.com/urfave/cli"

var (
	flags = []cli.Flag{
		cli.StringFlag{
			Name:   "vsphere-url",
			Usage:  "vSphere SDK URL, e.g. https://admin:password@192.0.2.1/sdk",
			EnvVar: "VSPHERE_MONITOR_VSPHERE_URL,VSPHERE_URL",
		},
		cli.BoolFlag{
			Name:   "vsphere-insecure",
			Usage:  "Whether to skip TLS certificate verification when connecting to the vSphere SDK",
			EnvVar: "VSPHERE_MONITOR_VSPHERE_INSECURE,VSPHERE_INSECURE",
		},
		cli.StringSliceFlag{
			Name:   "vsphere-cluster-path",
			Usage:  "Inventory path to the cluster to monitor. Can be specified multiple times for multiple clusters, or use comma-separated values if using the environment variable.",
			EnvVar: "VSPHERE_MONITOR_VSPHERE_CLUSTER_PATHS,VSPHERE_CLUSTER_PATHS",
		},
		cli.StringSliceFlag{
			Name:   "vsphere-host-alert-id-metric-name",
			Usage:  "alert-id:metric-name mappings. Can be specified multiple times for multiple alerts, or use comma-separated values if using the environment variable.",
			EnvVar: "VSPHERE_MONITOR_VSPHERE_HOST_ALERT_ID_METRIC_NAMES,VSPHERE_HOST_ALERT_ID_METRIC_NAMES",
		},
		cli.StringFlag{
			Name:   "librato-email",
			Usage:  "Email address for the Librato account to send metrics to",
			EnvVar: "VSPHERE_MONITOR_LIBRATO_EMAIL,LIBRATO_EMAIL",
		},
		cli.StringFlag{
			Name:   "librato-token",
			Usage:  "Librato token (with record permissions) associated with the Librato account to send metrics to",
			EnvVar: "VSPHERE_MONITOR_LIBRATO_TOKEN,LIBRATO_TOKEN",
		},
		cli.StringFlag{
			Name:   "librato-source-prefix",
			Usage:  "Prefix to prepent to all Librato source names",
			EnvVar: "VSPHERE_MONITOR_LIBRATO_SOURCE_PREFIX,LIBRATO_SOURCE_PREFIX",
		},
		cli.StringFlag{
			Name:   "sentry-dsn",
			Usage:  "Sentry DSN for sending errors to Sentry",
			EnvVar: "VSPHERE_MONITOR_SENTRY_DSN,SENTRY_DSN",
		},
		cli.StringFlag{
			Name:   "sentry-environment",
			Usage:  "Environment name to pass to Sentry when sending errors",
			EnvVar: "VSPHERE_MONITOR_SENTRY_ENVIRONMENT,SENTRY_ENVIRONMENT",
		},
	}
)
