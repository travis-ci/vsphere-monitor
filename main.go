package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
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

	alarmMap := make(map[string]mo.Alarm)

	for _, host := range hosts {
		fmt.Printf("%s:\n", host.Name())
		alarmStateResp, err := methods.GetAlarmState(ctx, client.Client, &types.GetAlarmState{This: *alarmManager, Entity: host.Reference()})
		if err != nil {
			fmt.Printf("\tcouldn't get alarm states: %v\n", err)
			continue
		}

		for _, state := range alarmStateResp.Returnval {
			alarm, ok := alarmMap[state.Alarm.Value]
			if !ok {
				err := client.RetrieveOne(ctx, state.Alarm, nil, &alarm)
				if err != nil {
					fmt.Printf("couldn't get alarm info for %v\n", state.Alarm)
					continue
				}
				alarmMap[state.Alarm.Value] = alarm
			}

			if state.OverallStatus == "gray" {
				continue
			}

			fmt.Printf("\t%v -> %s\n", alarm.Info.Name, state.OverallStatus)
		}
	}
}
