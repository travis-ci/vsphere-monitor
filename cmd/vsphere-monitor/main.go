package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	raven "github.com/getsentry/raven-go"
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

	app.Flags = flags
	app.Action = mainAction

	var err error
	raven.CapturePanicAndWait(func() {
		err = app.Run(os.Args)
	}, nil)
	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		fmt.Printf("%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func mainAction(c *cli.Context) error {
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	logger := logrus.WithField("pid", os.Getpid())

	logger.WithField("version", VersionString).Info("starting vsphere-monitor")
	defer logger.Info("stopping vsphere-monitor")

	if c.String("sentry-dsn") != "" {
		err := raven.SetDSN(c.String("sentry-dsn"))
		if err == nil {
			raven.SetRelease(VersionString)
			if c.String("sentry-environment") != "" {
				raven.SetEnvironment(c.String("sentry-environment"))
			}

			logger.Info("set up Sentry")
		} else {
			logger.WithError(err).Error("couldn't set Sentry DSN")
		}
	}

	vsphereURL, err := url.Parse(c.String("vsphere-url"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse vSphere URL")
	}

	ctx := context.Background()

	vSphereClient, err := vspheremonitor.NewVSphereClient(ctx, vsphereURL, c.Bool("vsphere-insecure"))
	if err != nil {
		return errors.Wrap(err, "error creating vsphere client")
	}

	har := vspheremonitor.HostAlarmReporter{
		LibratoClient:        vspheremonitor.NewLibratoClient(c.String("librato-email"), c.String("librato-token"), c.String("librato-source-prefix")),
		VSphereClient:        vSphereClient,
		AlarmIDMetricNameMap: kvSliceToMap(c.StringSlice("vsphere-host-alert-id-metric-name"), ":"),
	}

	err = har.SetClusterPaths(ctx, logger, c.StringSlice("vsphere-cluster-path"))
	if err != nil {
		return errors.Wrap(err, "error setting up clusters")
	}

	ticker := time.Tick(time.Minute)
	for range ticker {
		har.Report(ctx, logger)
	}

	return nil
}

func kvSliceToMap(kvs []string, separator string) map[string]string {
	kvMap := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		parts := strings.SplitN(kv, separator, 2)
		kvMap[parts[0]] = parts[1]
	}
	return kvMap
}
