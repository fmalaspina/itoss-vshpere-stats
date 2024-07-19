package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"
	"os"
	"strings"
)

func GetHostStats(ctx context.Context, err error, v *view.ContainerView, functions []string, entityToQuery string) error {

	hostName, err := getEntityName(ctx, v, entityToQuery, *entityNameFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting host name: %s\n", err)
		os.Exit(1)
	}
	var metricsToQuery []string

	if len(strings.Split(*metricsFlag, ",")) > 1 {
		metricsToQuery = strings.Split(*metricsFlag, ",")
	} else {
		metricsToQuery = []string{*metricsFlag}

	}

	// construct titles
	title := ""

	for range metricsToQuery {
		title += "entity;name;instance;metric"
		for _, function := range functions {
			title += ";" + function
		}
		title += ";units|"
	}
	// delete the las pipe character
	title = title[:len(title)-1]

	fmt.Println(title)

	vmsRefs, err := v.Find(ctx, []string{entityToQuery}, nil)
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
		MaxSample:  int32(*maxSamplesFlag),
		MetricId:   []types.PerfMetricId{{Instance: *instanceFlag}},
		IntervalId: int32(*intervalFlag),
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, metricsToQuery, vmsRefs)
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
		if *entityNameFlag != "all" && name.Value != hostName {
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
				resultLine += fmt.Sprintf("%s;%s;%s;%s",
					name.Type, *entityNameFlag, instance, v.Name)

				for _, function := range functions {
					result, err := applyFunction(values, function)
					if err != nil {
						fmt.Fprint(os.Stderr, "Error applying function:", err, "\n")
						os.Exit(1)
					}
					resultLine += fmt.Sprintf(";%.2f", result)
				}
				resultLine += fmt.Sprintf(";%s;", units)
				if *instanceFlag != "" {
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
		fmt.Fprintf(os.Stderr, "\nMetric not found for entity %s\n", *entityNameFlag)
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
