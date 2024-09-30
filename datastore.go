package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"os"
)

func GetDatastoreStatus(ctx context.Context, c *vim25.Client, err error) error {
	m := view.NewManager(c)
	vHost, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}
	defer vHost.Destroy(ctx)
	hostName, err := getEntityName(ctx, vHost, "HostSystem", *entityNameFlag)
	if err != nil {
		return err
	}

	vDatastore, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return err
	}
	defer vDatastore.Destroy(ctx)
	fmt.Fprint(os.Stdout, "name;host,internalHostname,type;maintenanceMode;multipleHostAccess;overallStatus;capacity;freeSpace;Uncommitted;accessible\n")
	//hosts := extractHostsMountedOn(ctx, v)
	var dss []mo.Datastore

	err = vDatastore.Retrieve(ctx, []string{"Datastore"}, []string{"summary", "host"}, &dss)

	for _, ds := range dss {

		for _, host := range ds.Host {
			if hostName == host.Key.Value {
				fmt.Fprintf(os.Stdout, "%s;%v;%s;%s;%v;%v;%v;%v;%v;%v,%v\n",
					ds.Summary.Name,
					*entityNameFlag,
					host.Key.Value,
					ds.Summary.Type,
					ds.Summary.MaintenanceMode,
					ds.Summary.MultipleHostAccess,
					ds.ManagedEntity.OverallStatus,
					ds.Summary.Capacity,
					ds.Summary.FreeSpace,
					ds.Summary.Uncommitted,
					ds.Summary.Accessible)
			}
		}
	}

	return nil
}

//func extractHostsMountedOn(ctx context.Context, v *view.ContainerView) []string {
//
//	hosts := []string{}
//
//	for _, ds := range dss {
//		for _, host := range ds.Host {
//			hostName, _ := getEntityName(ctx, v, "HostSystem", host.Key.Value)
//			hosts = append(hosts, hostName)
//		}
//	}
//	return hosts
//}
//
//func mountedOnHost(ctx context.Context, v *view.ContainerView, hosts []string) (bool, error) {
//	var hss []mo.HostSystem
//
//	err := v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
//	if err != nil {
//		return false, err
//	}
//	//for  := range hss {
//	//	hostName, err := getEntityName(ctx, v, "HostSystem", hs.Summary.Config.Name)
//	//}
//	return true, nil
//}
