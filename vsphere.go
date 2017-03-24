package vspheremonitor

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
)

// A VSphereClient allows for communicating with the vSphere SDK API.
type VSphereClient struct {
	client *govmomi.Client
}

// NewVSphereClient creates a new vSphere Client, connects to the API and
// authenticates with it to set up a session. The URL should be the URL to the
// SDK endpoint, and should include the username and password, for example
// "https://admin:password@192.0.2.1/sdk". The insecure argument determines
// whether the TLS certificate of the SDK endpoint should be verified or not.
func NewVSphereClient(ctx context.Context, url *url.URL, insecure bool) (*VSphereClient, error) {
	client, err := govmomi.NewClient(ctx, url, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "error creating govmomi client")
	}

	return &VSphereClient{client: client}, nil
}

// ListHostsInCluster returns a list of hosts in the cluster at the given inventory path. Note that the cluster path should be a full inventory path, for example "/datacenter-name/host/cluster-name". If an error occurs, a nil slice is returned with the error.
func (vc *VSphereClient) ListHostsInCluster(ctx context.Context, clusterPath string) ([]*VSphereHost, error) {
	finder := find.NewFinder(vc.client.Client, true)
	govmomiHosts, err := finder.HostSystemList(ctx, clusterPath)
	if err != nil {
		return nil, errors.Wrap(err, "error listing hosts")
	}

	hosts := make([]*VSphereHost, 0, len(govmomiHosts))
	for _, govmomiHost := range govmomiHosts {
		hosts = append(hosts, &VSphereHost{hostSystem: govmomiHost})
	}

	return hosts, nil
}

// ListAlarmStatesForHost returns a map with the alarm IDs of the alarms defined on the given host as keys, and their current status as the value. Valid statuses are "green", "yellow", "red" and "gray". If an error occurs, it is returned and the map is nil.
func (vc *VSphereClient) ListAlarmStatesForHost(ctx context.Context, host *VSphereHost) (map[string]string, error) {
	alarmManager := vc.client.Client.ServiceContent.AlarmManager
	if alarmManager == nil {
		return nil, errors.New("client has no alarm manager")
	}

	alarmStateResp, err := methods.GetAlarmState(ctx, vc.client.Client, &types.GetAlarmState{
		This:   *alarmManager,
		Entity: host.hostSystem.Reference(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error fetching alarm states")
	}

	alarmStates := make(map[string]string)
	for _, state := range alarmStateResp.Returnval {
		alarmStates[state.Alarm.Value] = string(state.OverallStatus)
	}

	return alarmStates, nil
}

// A VSphereHost represents a single ESXi host
type VSphereHost struct {
	hostSystem *object.HostSystem
}

// Name returns the inventory name of the VSphereHost
func (vh *VSphereHost) Name() string {
	return vh.hostSystem.Name()
}
