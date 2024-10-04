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

func GetHostsConfig(ctx context.Context, c *vim25.Client) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem

	err = v.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary", "hardware"}, &hss, property.Match{"name": *entityNameFlag})

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

	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary", "config", "layout", "resourcePool", "parent", "snapshot"}, &vms, property.Match{"name": *entityNameFlag})

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
