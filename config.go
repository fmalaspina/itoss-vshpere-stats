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

func GetHostsConfig(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem

	err = v.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary", "hardware"}, &hss, property.Match{"name": hostFlag})

	if err != nil {
		return err
	}
	hostFound := false

	fmt.Fprint(os.Stdout, "host;vendor;model;memorySize;cpuModel;cpuMhz;numCpuCores;numCpuThreads;fullName;version;build;patchLevel\n")
	for _, hs := range hss {
		//if *entityNameFlag != "all" && hs.Summary.Config.Name != *entityNameFlag {
		//	continue
		//}

		fmt.Fprintf(os.Stdout, "%s;%s;%s;%v;%s;%v;%v;%v;%s;%s;%s;%s\n",
			safeValue(hs.Summary.Config.Name),
			safeValue(hs.Summary.Hardware.Vendor),
			safeValue(hs.Summary.Hardware.Model),
			safeValue(hs.Summary.Hardware.MemorySize),
			safeValue(hs.Summary.Hardware.CpuModel),
			safeValue(hs.Summary.Hardware.CpuMhz),
			safeValue(hs.Summary.Hardware.NumCpuCores),
			safeValue(hs.Summary.Hardware.NumCpuThreads),
			safeValue(hs.Summary.Config.Product.FullName),
			safeValue(hs.Summary.Config.Product.Version),
			safeValue(hs.Summary.Config.Product.Build),
			safeValue(hs.Summary.Config.Product.PatchLevel))

		//
		hostFound = true
	}
	if !hostFound {
		//fmt.Fprintf(os.Stdout, "%s;%s;%s;%v;%s;%v;%v;%v;%s;%s;%s;%s\n",
		//	*entityNameFlag, "NA", "NA", "NA", "NA", "NA", "NA", "NA", "NA", "NA", "NA", "HOST_NOT_FOUND")
		//os.Exit(0)
		fmt.Fprintf(os.Stderr, "\nError: %s\n", "Host not found.")
		os.Exit(1)
	}
	return nil
}

func GetVMConfig(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine

	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary", "config", "layout", "resourcePool", "parent", "snapshot"}, &vms, property.Match{"name": vmFlag})

	if err != nil {
		return err
	}
	vmFound := false

	fmt.Fprint(os.Stdout, "name;internalName;numEthernetCards;numVirtualDisks;hwVersion;memorySizeMB;memoryReservation;numCpu;cpuReservation;guestFullName\n")
	for _, vm := range vms {
		//if *entityNameFlag != "all" && vm.Summary.Config.Name != *entityNameFlag {
		//	continue
		//}

		fmt.Fprintf(os.Stdout, "%s;%s;%v;%v;%s;%v;%v;%v;%v;%s\n",
			safeValue(vm.Summary.Config.Name),
			safeValue(vm.Summary.Vm.Value),
			safeValue(vm.Summary.Config.NumEthernetCards),
			safeValue(vm.Summary.Config.NumVirtualDisks),
			safeValue(vm.Summary.Config.HwVersion),
			safeValue(vm.Summary.Config.MemorySizeMB),
			safeValue(vm.Summary.Config.MemoryReservation),
			safeValue(vm.Summary.Config.NumCpu),
			safeValue(vm.Summary.Config.CpuReservation),
			safeValue(vm.Summary.Config.GuestFullName),
		)

		//
		vmFound = true
	}
	if !vmFound {
		//fmt.Fprintf(os.Stdout, "%s;%s;%s;%s;%s;%s\n",
		//	*entityNameFlag, "NA", "NA", "NA", "NA", "NA")
		//os.Exit(0)
		fmt.Fprintf(os.Stderr, "\nError: %s\n", "VM not found.")
		os.Exit(1)
	}
	return nil
}

func GetClusterConfig(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var clusters []mo.ClusterComputeResource

	err = v.RetrieveWithFilter(ctx, []string{"ClusterComputeResource"}, []string{"summary", "configuration", "host", "datastore"}, &clusters, property.Match{"self.value": clusterFlag})

	if err != nil {
		return err
	}

	var hostNames, datastoreNames strings.Builder

	clusterFound := false

	fmt.Fprint(os.Stdout, "name;hosts;datastores;totalCpu;totalMemory;numCpuCores;numCpuThreads;effectiveCpu;effectiveMemory;numHosts;numEffectiveHosts\n")

	for _, cluster := range clusters {
		hostNames.Reset()
		datastoreNames.Reset()
		for i, host := range cluster.Host {
			if i > 0 {
				hostNames.WriteString(",")
			}
			hostNames.WriteString(host.Value)
		}
		for i, datastore := range cluster.Datastore {
			if i > 0 {
				datastoreNames.WriteString(",")
			}
			datastoreNames.WriteString(datastore.Value)
		}
		fmt.Fprintf(os.Stdout, "%s;%s;%s;%v;%d;%v;%v;%v;%d;%v;%v\n",
			safeValue(cluster.Self.Value),
			safeValue(hostNames.String()),
			safeValue(datastoreNames.String()),
			safeValue(cluster.Summary.GetComputeResourceSummary().TotalCpu),
			safeValue(cluster.Summary.GetComputeResourceSummary().TotalMemory),
			safeValue(cluster.Summary.GetComputeResourceSummary().NumCpuCores),
			safeValue(cluster.Summary.GetComputeResourceSummary().NumCpuThreads),
			safeValue(cluster.Summary.GetComputeResourceSummary().EffectiveCpu),
			safeValue(cluster.Summary.GetComputeResourceSummary().EffectiveMemory),
			safeValue(cluster.Summary.GetComputeResourceSummary().NumHosts),
			safeValue(cluster.Summary.GetComputeResourceSummary().NumEffectiveHosts),
		)
		clusterFound = true
	}

	if !clusterFound {
		fmt.Fprintf(os.Stderr, "\nError: %s\n", "Cluster not found.")
		os.Exit(1)
	}
	return nil
}

func GetResourcePoolConfig(ctx context.Context, c *vim25.Client) error {
	// based on GetClusterConfig
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ResourcePool"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var resourcePools []mo.ResourcePool

	err = v.RetrieveWithFilter(ctx, []string{"ResourcePool"}, []string{"parent", "namespace", "name", "summary", "owner", "config", "vm", "runtime"}, &resourcePools, property.Match{"self.value": resourcePoolFlag})

	if err != nil {
		return err

	}
	var vmNames strings.Builder

	fmt.Fprint(os.Stdout, "name;vmNames;cpuReservation;cpuExpandableReservation;cpuLimit;memoryReservation;memoryExpandableReservation;memoryLimit\n")

	resourcePoolFound := false

	for _, rp := range resourcePools {
		vmNames.Reset()

		for i, vmName := range rp.Vm {
			if i > 0 {
				vmNames.WriteString(",")
			}
			vmNames.WriteString(vmName.Value)
		}

		fmt.Fprintf(os.Stdout, "%s;%s;%s;%d;%t;%d;%d;%t;%d\n",
			safeValue(rp.Self.Value),
			safeValue(rp.Parent.Value),
			safeValue(vmNames.String()),
			safeValue(*rp.Config.CpuAllocation.Reservation),
			safeValue(*rp.Config.CpuAllocation.ExpandableReservation),
			safeValue(*rp.Config.CpuAllocation.Limit),
			safeValue(*rp.Config.MemoryAllocation.Reservation),
			safeValue(*rp.Config.MemoryAllocation.ExpandableReservation),
			safeValue(*rp.Config.MemoryAllocation.Limit),
		)

		//
		resourcePoolFound = true
	}
	if !resourcePoolFound {
		fmt.Fprintf(os.Stderr, "\nError: %s\n", "Resource pool not found.")
		os.Exit(1)
	}
	return nil
}
