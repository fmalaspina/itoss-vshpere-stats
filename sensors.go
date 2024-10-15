package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"os"
)

// TODO revisar que no se comporte como host status debe ser fault
func GetHostsSensors(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem

	err = v.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary", "runtime"}, &hss, property.Match{"name": hostFlag})

	if err != nil {
		return err
	}
	hostFound := false

	fmt.Fprint(os.Stdout, "host;name;key;currentReading;unitModifier;BaseUnits;sensorType;id;timestamp\n")
	for _, hs := range hss {
		//if *entityNameFlag != "all" && hs.Summary.Config.Name != *entityNameFlag {
		//	continue
		//}
		sensorsFound := false
		for _, sensor := range hs.Runtime.HealthSystemRuntime.SystemHealthInfo.NumericSensorInfo {
			fmt.Fprintf(os.Stdout, "%s;%s;%v;%v;%s;%s;%v;%s\n",
				safeValue(hs.Summary.Config.Name),
				safeValue(sensor.Name),
				safeValue(sensor.HealthState.GetElementDescription().Key),
				safeValue(sensor.CurrentReading),
				safeValue(sensor.BaseUnits),
				safeValue(sensor.SensorType),
				safeValue(sensor.Id),
				safeValue(sensor.TimeStamp))
			sensorsFound = true
		}
		if !sensorsFound {
			fmt.Fprintf(os.Stderr, "Error getting sensors from host %s!\n", hostFlag)
			os.Exit(1)
		}
		//
		hostFound = true
	}
	if !hostFound {

		fmt.Fprintf(os.Stderr, "Host %s not found!\n", hostFlag)
		os.Exit(1)

	}
	return nil
}
