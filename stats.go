package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"os"
	"strings"
)

func GetHostStats(ctx context.Context, c *vim25.Client, functions []string) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary"}, &hss, property.Match{"name": hostFlag})
	//if err != nil {
	//	return err
	//}
	//	hostName, err := getHostNames(ctx, v, *entityNameFlag)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting host name: %s\n", err)
		os.Exit(1)
	}

	var hostNames []string

	var internalHostnames = make(map[string]string)
	// Iterate over the host systems and collect names
	for _, hs := range hss {
		hostNames = append(hostNames, hs.Summary.Host.Value)
		internalHostnames[hs.Summary.Host.Value] = hs.Summary.Config.Name
	}
	return getStats(ctx, err, v, functions, "HostSystem", hostNames, internalHostnames, hostFlag)

}

func GetVMStats(ctx context.Context, c *vim25.Client, functions []string) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms, property.Match{"name": vmFlag})
	var vmNames []string
	var internalVMNames = make(map[string]string)
	// Iterate over the host systems and collect names
	for _, vm := range vms {

		vmNames = append(vmNames, vm.Summary.Vm.Value)
		internalVMNames[vm.Summary.Vm.Value] = vm.Summary.Config.Name
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting vm name: %s\n", err)
		os.Exit(1)
	}
	return getStats(ctx, err, v, functions, "VirtualMachine", vmNames, internalVMNames, vmFlag)

}

func GetResourcePoolStats(ctx context.Context, c *vim25.Client, functions []string) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ResourcePool"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var rp []mo.ResourcePool
	err = v.RetrieveWithFilter(ctx, []string{"ResourcePool"}, []string{"name", "summary", "config", "runtime"}, &rp, property.Match{"self.value": resourcePoolFlag})
	var rpNames []string
	var internalRPNames = make(map[string]string)
	// Iterate over the host systems and collect names
	for _, r := range rp {

		rpNames = append(rpNames, r.Self.Value)
		internalRPNames[r.Self.Value] = r.Self.Value
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting resource pool name: %s\n", err)
		os.Exit(1)
	}
	return getStats(ctx, err, v, functions, "ResourcePool", rpNames, internalRPNames, resourcePoolFlag)

}

func GetClusterStats(ctx context.Context, c *vim25.Client, functions []string) error {
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"ClusterComputeResource"}, true)
	if err != nil {
		return err
	}
	defer v.Destroy(ctx)

	var cr []mo.ClusterComputeResource
	err = v.RetrieveWithFilter(ctx, []string{"ClusterComputeResource"}, []string{"summary"}, &cr, property.Match{"self.value": clusterFlag})
	var crNames []string
	var internalCRNames = make(map[string]string)
	// Iterate over the host systems and collect names
	for _, c := range cr {

		crNames = append(crNames, c.Self.Value)
		internalCRNames[c.Self.Value] = c.Self.Value
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cluster name: %s\n", err)
		os.Exit(1)
	}
	return getStats(ctx, err, v, functions, "ClusterComputeResource", crNames, internalCRNames, clusterFlag)

}

func getStats(ctx context.Context, err error, v *view.ContainerView, functions []string, entityToQuery string, names []string, internalNames map[string]string, flag string) error {
	var metricsToQuery []string

	if len(strings.Split(metricsFlag, ",")) > 1 {
		metricsToQuery = strings.Split(metricsFlag, ",")
	} else {
		metricsToQuery = []string{metricsFlag}

	}

	// construct titles
	title := ""

	for range metricsToQuery {
		title += "entity;name;internalName;instance;metric"
		for _, function := range functions {
			title += ";" + function
		}
		title += ";units|"
	}
	// delete the las pipe character
	title = title[:len(title)-1]

	fmt.Println(title)

	entityRefs, err := v.Find(ctx, []string{entityToQuery}, nil)
	if err != nil {
		return err
	}

	// Create a PerfManager
	perfManager := performance.NewManager(v.Client())

	// Retrieve counters name list
	counters, err := perfManager.CounterInfoByName(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting counters: %s\n", err)
		os.Exit(1)
	}

	// Check if the metrics to query exist
	err = checkMetricExistence(counters, metricsToQuery)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	// Create PerfQuerySpec
	spec := types.PerfQuerySpec{
		MaxSample:  int32(maxSamplesFlag),
		MetricId:   []types.PerfMetricId{{Instance: instanceFlag}},
		IntervalId: int32(intervalFlag),
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, metricsToQuery, entityRefs)
	if err != nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting metric: %s\n", err)
			os.Exit(1)
		}
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting metric series: %s\n", err)
			os.Exit(1)
		}
	}

	// Read result
	var results []string
	for _, metric := range result {
		resultLine := ""
		name := metric.Entity
		if flag != "*" && !contains(names, name.Value) {
			continue
		}
		for _, v := range metric.Value {
			counter := counters[v.Name]
			units := counter.UnitInfo.GetElementDescription().Label

			instance := v.Instance
			if instance == "" {
				instance = "-"
			}

			if len(v.Value) != 0 {
				values, err := parseCSV(v.ValueCSV())
				if err != nil {
					fmt.Fprint(os.Stderr, "Error parsing metric CSV values: ", err, "\n")
					os.Exit(1)
				}
				//fmt.Printf("entity=%s;name=%s;instance=%s;metric=%s",
				//	name.Type, name.Value, instance, v.Name)
				resultLine += fmt.Sprintf("%s;%s;%s;%s;%s",
					name.Type, internalNames[name.Value], name.Value, instance, v.Name)

				for _, function := range functions {
					result, err := applyFunction(values, function)
					if err != nil {
						fmt.Fprint(os.Stderr, "Error applying function:", err, "\n")
						os.Exit(1)
					}
					resultLine += fmt.Sprintf(";%.2f", result)
				}
				resultLine += fmt.Sprintf(";%s;", units)
				if instanceFlag != "" {
					resultLine += "\n"
				} else {
					resultLine += "|"
				}

			} else {
				fmt.Fprintf(os.Stderr, "No values found for metric %s\n", v.Name)
				os.Exit(1)
			}

		}
		// delete last semicolon character
		resultLine = resultLine[:len(resultLine)-1]

		results = append(results, resultLine)

	}
	// print title and then results

	metricFound := false
	for _, result := range results {

		metricFound = true
		fmt.Println(result)
	}
	if !metricFound {
		fmt.Fprintf(os.Stderr, "\nMetric not found for entity %s\n", flag)
		os.Exit(1)
	}
	return nil
}

func checkMetricExistence(counterMap map[string]*types.PerfCounterInfo, metricNames []string) error {
	for _, key := range metricNames {
		if _, exists := counterMap[key]; !exists {
			return errors.New(fmt.Sprintf("Metric '%s' does not exist", key))
		}
	}
	return nil
}
