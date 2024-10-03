package main

import (
	"context"
	"fmt"
	"github.com/kr/pretty"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"os"
)

func GetClusterStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var ccr []mo.ClusterComputeResource

	err = v.RetrieveWithFilter(ctx, []string{"ClusterComputeResource"}, []string{"computeResource", "managedEntity", "summary"}, &ccr, property.Match{"name": *entityNameFlag})
	pretty.Print(ccr)
	return nil

}

func GetHostsStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	vHost, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}
	defer vHost.Destroy(ctx)

	var hss []mo.HostSystem

	err = vHost.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary"}, &hss, property.Match{"name": *entityNameFlag})

	if err != nil {
		return err
	}
	hostFound := false

	fmt.Fprint(os.Stdout, "host;uptimeSec;overallStatus;connectionState;inMaintenanceMode;powerState;standbyMode;bootTime;proxyStatus\n")
	for _, hs := range hss {
		//if *entityNameFlag != "all" && hs.Summary.Config.Name != *entityNameFlag {
		//	continue
		//}

		fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
			safeValue(hs.Summary.Config.Name),
			safeValue(hs.Summary.QuickStats.Uptime),
			safeValue(hs.Summary.OverallStatus),
			safeValue(hs.Summary.Runtime.ConnectionState),
			safeValue(hs.Summary.Runtime.InMaintenanceMode),
			safeValue(hs.Summary.Runtime.PowerState),
			safeValue(hs.Summary.Runtime.StandbyMode),
			safeValue(hs.Summary.Runtime.BootTime),
			"OK")
		//
		hostFound = true
	}
	if !hostFound {
		fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
			*entityNameFlag, 0, "NA", "NA", false, "NA", "NA", "NA", "HOST_NOT_FOUND")
		os.Exit(0)
	}
	return nil
}

func GetVMStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine

	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms, property.Match{"name": *entityNameFlag})

	if err != nil {
		return err
	}

	vmFound := false

	fmt.Fprint(os.Stdout, "name;internalName;overallStatus;connectionState;powerState;guestHeartbeatStatus;bootTime;uptimeSeconds;proxyStatus\n")
	for _, vm := range vms {
		//if *entityNameFlag != "all" && vm.Summary.Config.Name != *entityNameFlag {
		//	continue
		//}

		fmt.Fprintf(os.Stdout, "%s;%s;%s;%s;%s;%s;%s;%v;%s\n",
			safeValue(vm.Summary.Config.Name),
			safeValue(vm.Summary.Vm.Value),
			safeValue(vm.Summary.OverallStatus),
			safeValue(vm.Summary.Runtime.ConnectionState),
			safeValue(vm.Summary.Runtime.PowerState),
			safeValue(vm.Summary.QuickStats.GuestHeartbeatStatus),
			safeValue(vm.Summary.Runtime.BootTime),
			safeValue(vm.Summary.QuickStats.UptimeSeconds),
			"OK")

		vmFound = true
	}
	if !vmFound {
		fmt.Fprintf(os.Stdout, "%s;%s;%s;%s;%s;%s;%s;%v;%s\n",
			*entityNameFlag, "NA", "NA", "NA", "NA", "NA", "NA", 0, "VM_NOT_FOUND")
		os.Exit(0)
	}
	return nil
}

func GetDatastoreStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)

	vDatastore, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return err
	}
	defer vDatastore.Destroy(ctx)

	fmt.Fprint(os.Stdout, "name;type;maintenanceMode;capacity;freeSpace;uncommitted;accessible\n")

	// Destination slice to hold the result
	var dss []mo.Datastore

	// Retrieve datastores the match the filter
	err = vDatastore.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"summary", "host", "info", "vm"}, &dss, property.Match{"name": *entityNameFlag})

	if err != nil {
		return err
	}

	// Iterate over the filtered datastores
	for _, ds := range dss {
		fmt.Fprintf(os.Stdout, "%s;%s;%v;%v;%v;%v;%v\n",
			safeValue(ds.Summary.Name),
			safeValue(ds.Summary.Type),
			safeValue(ds.Summary.MaintenanceMode),
			safeValue(ds.Summary.Capacity),
			safeValue(ds.Summary.FreeSpace),
			safeValue(ds.Summary.Uncommitted),
			safeValue(ds.Summary.Accessible))
	}

	return nil
}
