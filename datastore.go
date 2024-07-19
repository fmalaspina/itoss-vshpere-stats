package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"os"
)

func GetDatastoreStatus(ctx context.Context, err error, v *view.ContainerView) error {

	fmt.Fprint(os.Stdout, "name;type;maintenanceMode;multipleHostAccess;overallStatus;capacity;freeSpace;Uncommitted;accessible\n")
	//hosts := extractHostsMountedOn(ctx, v)
	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary", "host"}, &dss)

	if err != nil {
		return err
	}
	for _, ds := range dss {
		//found, err := mountedOnHost(ctx, v, hosts)
		if err != nil {
			return err
		}
		//if found {
		fmt.Fprintf(os.Stdout, "%s;%s;%s;%v;%v;%v;%v;%v;%v\n",
			ds.Summary.Name,
			ds.Summary.Type,
			ds.Summary.MaintenanceMode,
			ds.Summary.MultipleHostAccess,
			ds.ManagedEntity.OverallStatus,
			ds.Summary.Capacity,
			ds.Summary.FreeSpace,
			ds.Summary.Uncommitted,
			ds.Summary.Accessible)
		//}
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
