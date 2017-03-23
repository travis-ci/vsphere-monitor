package main

import "github.com/urfave/cli"

var (
	Flags = []cli.Flag{
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
		cli.StringFlag{
			Name:   "vsphere-cluster-path",
			Usage:  "Inventory path to the cluster to monitor",
			EnvVar: "VSPHERE_MONITOR_VSPHERE_CLUSTER_PATH,VSPHERE_CLUSTER_PATH",
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
	}
)
