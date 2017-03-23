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

type VSphereClient struct {
	client *govmomi.Client
}

func NewVSphereClient(ctx context.Context, url *url.URL, insecure bool) (*VSphereClient, error) {
	client, err := govmomi.NewClient(ctx, url, insecure)
	if err != nil {
		return nil, errors.Wrap(err, "error creating govmomi client")
	}

	return &VSphereClient{client: client}, nil
}

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

type VSphereHost struct {
	hostSystem *object.HostSystem
}

func (vh *VSphereHost) Name() string {
	return vh.hostSystem.Name()
}
