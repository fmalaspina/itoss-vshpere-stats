package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"os"
	"strings"
)

func GetClusterStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		showClusterError(err.Error())
	}
	defer v.Destroy(ctx)
	var ccr []mo.ClusterComputeResource

	err = v.RetrieveWithFilter(ctx, []string{"ClusterComputeResource"}, []string{"computeResource", "managedEntity", "summary", "extensibleManagedObject"}, &ccr, property.Match{"self.value": clusterFlag})
	if err != nil {
		showClusterError(err.Error())
	}
	clusterFound := false

	fmt.Fprint(os.Stdout, "cluster;totalCpu;totalMemory;numCpuCores;numCpuThreads;effectiveCpu;effectiveMemory;numHosts;numEffectiveHosts;overallStatus\n")
	for _, cr := range ccr {

		fmt.Fprintf(os.Stdout, "%s;%d;%d;%d;%d;%d;%d;%d;%d;%s;%s\n",
			safeValue(cr.ManagedEntity.ExtensibleManagedObject.Self.Value),
			safeValue(cr.Summary.GetComputeResourceSummary().TotalCpu),
			safeValue(cr.Summary.GetComputeResourceSummary().TotalMemory),
			safeValue(cr.Summary.GetComputeResourceSummary().NumCpuCores),
			safeValue(cr.Summary.GetComputeResourceSummary().NumCpuThreads),
			safeValue(cr.Summary.GetComputeResourceSummary().EffectiveCpu),
			safeValue(cr.Summary.GetComputeResourceSummary().EffectiveMemory),
			safeValue(cr.Summary.GetComputeResourceSummary().NumHosts),
			safeValue(cr.Summary.GetComputeResourceSummary().NumEffectiveHosts),
			safeValue(cr.Summary.GetComputeResourceSummary().OverallStatus),
			"OK")

		clusterFound = true
	}
	if !clusterFound {
		//fmt.Fprintf(os.Stdout, "%s;%d;%d;%d;%d;%d;%d;%d;%d;%s;%s\n",
		//	*entityNameFlag, 0, 0, 0, 0, 0, 0, 0, 0, "NA", "CLUSTER_NOT_FOUND")
		//os.Exit(0)
		showClusterError("CLUSTER_NOT_FOUND")
	}
	return nil

}

func GetHostsStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	vHost, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		showHostStatusError(err.Error())
	}
	defer vHost.Destroy(ctx)

	var hss []mo.HostSystem

	err = vHost.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary"}, &hss, property.Match{"name": hostFlag})

	if err != nil {
		showHostStatusError(err.Error())
	}
	hostFound := false

	fmt.Fprint(os.Stdout, "host;uptimeSec;overallStatus;connectionState;inMaintenanceMode;powerState;standbyMode;bootTime;proxyStatus\n")
	for _, hs := range hss {

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
		//fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
		//	*entityNameFlag, 0, "NA", "NA", false, "NA", "NA", "NA", "HOST_NOT_FOUND")
		//os.Exit(0)
		showHostStatusError("HOST_NOT_FOUND")
	}
	return nil
}

func GetVMStatus(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		showVMStatusError(err.Error())
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine

	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms, property.Match{"name": vmFlag})

	if err != nil {
		showVMStatusError(err.Error())
	}

	vmFound := false

	fmt.Fprint(os.Stdout, "name;internalName;overallStatus;connectionState;powerState;guestHeartbeatStatus;bootTime;uptimeSeconds;proxyStatus\n")
	for _, vm := range vms {

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
		//fmt.Fprintf(os.Stdout, "%s;%s;%s;%s;%s;%s;%s;%v;%s\n",
		//	*entityNameFlag, "NA", "NA", "NA", "NA", "NA", "NA", 0, "VM_NOT_FOUND")
		//os.Exit(0)
		showVMStatusError("VM_NOT_FOUND")
	}
	return nil
}

func GetDatastoreStatus(ctx context.Context, c *vim25.Client) error {

	// getHostNames using hostedByFlag
	//var hostName string
	var err error
	var hostNames []string
	if hostFlag != "*" {
		hostNames, err = getHostNames(ctx, c, hostFlag)
	}

	if err != nil {
		showHostStatusError(err.Error())
	}

	m := view.NewManager(c)

	vDatastore, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		showDatastoreStatusError(err.Error())
	}
	defer vDatastore.Destroy(ctx)

	fmt.Fprint(os.Stdout, "name;type;maintenanceMode;capacity;freeSpace;uncommitted;accessible;hostedBy;hostedByInternal\n")

	// Destination slice to hold the result
	var dss []mo.Datastore

	// Retrieve datastores the match the filter
	err = vDatastore.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"summary", "host", "info", "vm"}, &dss, property.Match{"name": datastoreFlag})

	if err != nil {
		showDatastoreStatusError(err.Error())
	}
	datastoreFound := false
	// Iterate over the filtered datastores
	for _, ds := range dss {
		// If the datastore is not hosted by the host, skip it
		var internalHostValues []string
		for _, host := range ds.Host {
			if hostFlag != "*" {
				if !contains(hostNames, host.Key.Value) {
					continue

				}
			}
			internalHostValues = append(internalHostValues, host.Key.Value)
		}
		fmt.Fprintf(os.Stdout, "%s;%s;%v;%v;%v;%v;%v;%s;%s\n",
			safeValue(ds.Summary.Name),
			safeValue(ds.Summary.Type),
			safeValue(ds.Summary.MaintenanceMode),
			safeValue(ds.Summary.Capacity),
			safeValue(ds.Summary.FreeSpace),
			safeValue(ds.Summary.Uncommitted),
			safeValue(ds.Summary.Accessible),
			safeValue(hostFlag),
			safeValue(strings.Join(internalHostValues, ",")))

		datastoreFound = true
	}
	if !datastoreFound {
		showDatastoreStatusError("DATASTORE_NOT_FOUND")
	}
	return nil
}
