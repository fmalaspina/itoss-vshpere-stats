package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"os"
	"strconv"
	"strings"
	"time"
)

var urlFlag = flag.String("url", "", "Required. Usage: -url <https://username:password@host/sdk> (domain users can be set as username@domain)")
var insecureFlag = flag.Bool("insecure", false, "Required. Usage: -insecure")
var entityFlag = flag.String("entity", "host", "Optional. Usage: -entity <host|vm|resourcepool>")
var contextFlag = flag.String("context", "status", "Optional. Usage: -context <status|config|metrics>")
var entityNameFlag = flag.String("entityName", "all", "Optional. Usage: -entityName <host name| vm name")
var timeoutFlag = flag.Duration("timeout", 10*time.Second, "Optional. Usage: -timeout <timeout in duration Ex.: 10s (ms,h,m can be used as well)>")
var intervalFlag = flag.Int("i", 0, "Optional. Usage: -i <interval id>")
var metricsFlag = flag.String("metrics", "cpu.usage.average", "Optional. Usage: -metrics <cpu.usage.average, mem.usage.average>")
var functionsFlag = flag.String("functions", "last", "Optional. Usage: -functions <min,max,avg,last>")
var maxSamplesFlag = flag.Int("maxSamples", 1, "Optional. Usage: -maxSamples <number of samples>")
var instanceFlag = flag.String("instance", "*", "Optional. Usage: -instance <instance name>")

// NewClient creates a vim25.Client for use in the examples
func NewClient(ctx context.Context) (*vim25.Client, error) {
	// Parse URL from string

	u, err := soap.ParseURL(*urlFlag)

	if err != nil {
		return nil, err
	}

	s := &cache.Session{
		URL:      u,
		Insecure: *insecureFlag,
	}

	c := new(vim25.Client)
	err = s.Login(ctx, c, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Run calls f with Client create from the -url flag if provided,
// otherwise runs the example against vcsim.
func Run(f func(context.Context, *vim25.Client) error) {

	flag.Parse()
	var err error
	var c *vim25.Client

	if *contextFlag != "status" && *contextFlag != "metrics" {
		fmt.Fprint(os.Stdout, "Option not implemented, set context to status or metric.\n")
		flag.Usage()
		os.Exit(0)
	}
	if *entityFlag != "host" && *entityFlag != "vm" {
		fmt.Fprint(os.Stdout, "Option not implemented, set entity to host or vm.\n")
		flag.Usage()
		os.Exit(0)
	}
	if *urlFlag == "simulator" {
		err = simulator.VPX().Run(f)
		os.Exit(0)
	} else {
		if *urlFlag == "" {
			fmt.Fprint(os.Stdout, "You must specify an url.\n")
			flag.Usage()
			os.Exit(0)
		}
	}
	ctx := context.Background()

	ctx, _ = context.WithTimeout(ctx, *timeoutFlag)
	c, err = NewClient(ctx)
	errorText := ""
	if errors.Is(err, context.DeadlineExceeded) {
		errorText = "TIMEOUT"
	} else {
		errorText = "UNABLE_TO_CONNECT"
	}
	if err == nil {
		err = f(ctx, c)
	}

	if err != nil && *contextFlag == "status" {
		fmt.Fprint(os.Stdout, "host;uptimeSec;overallStatus;connectionState;inMaintenanceMode;powerState;standbyMode;bootTime;proxyStatus\n")
		fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
			"NA", 0, "NA", "NA", false, "NA", "NA", "NA", errorText)
		fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
		os.Exit(0)
	}
}

func main() {
	Run(func(ctx context.Context, c *vim25.Client) error {
		// Create a view of HostSystem objects
		m := view.NewManager(c)
		var entityToQuery = ""
		if *entityFlag == "host" {
			entityToQuery = "HostSystem"
		}
		if *entityFlag == "vm" {
			entityToQuery = "VirtualMachine"
		}
		if *contextFlag == "status" {
			v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{entityToQuery}, true)
			if err != nil {
				return err
			}
			defer v.Destroy(ctx)
			return GetHostsStatus(ctx, err, v, entityToQuery)
		} else if *contextFlag == "metrics" {
			if *metricsFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify metrics to query.\n")
				flag.Usage()
				os.Exit(0)
			}

			if *metricsFlag != "" {
				metrics := strings.Split(*metricsFlag, ",")
				if len(metrics) > 1 && *instanceFlag != "*" {
					fmt.Fprint(os.Stdout, "You must specify only one metric when using instance.\n")
					flag.Usage()
					os.Exit(0)
				}
			}
			var functions []string
			if *functionsFlag != "last" {
				functions = strings.Split(*functionsFlag, ",")

				for _, f := range functions {
					if !strings.Contains("last,min,max,avg", f) {
						fmt.Fprint(os.Stdout, "You must specify a valid function (avg,min,max,last).\n")
						flag.Usage()
						os.Exit(0)
					}
				}
			} else {
				functions = []string{"last"}
			}

			v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{entityToQuery}, true)
			if err != nil {
				return err
			}
			defer v.Destroy(ctx)
			return GetHostMetrics(ctx, err, v, functions, entityToQuery)
		}
		fmt.Fprint(os.Stdout, "Option not implemented. Set host status or host metrics.\n")
		flag.Usage()
		os.Exit(0)
		return nil
	})
}

func GetHostMetrics(ctx context.Context, err error, v *view.ContainerView, functions []string, entityToQuery string) error {

	vmsRefs, err := v.Find(ctx, []string{entityToQuery}, nil)
	if err != nil {
		return err
	}

	// Create a PerfManager
	perfManager := performance.NewManager(v.Client())

	// Retrieve counters name list
	counters, err := perfManager.CounterInfoByName(ctx)
	if err != nil {
		return err
	}

	var metricsToQuery []string

	if len(strings.Split(*metricsFlag, ",")) > 1 {
		metricsToQuery = strings.Split(*metricsFlag, ",")
	} else {
		metricsToQuery = []string{*metricsFlag}

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
		return err
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return err
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

	// Read result
	var results []string
	for _, metric := range result {
		resultLine := ""
		name := metric.Entity
		if *entityNameFlag != "all" && name.Value != *entityNameFlag {
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
					os.Exit(0)
				}
				//fmt.Printf("entity=%s;name=%s;instance=%s;metric=%s",
				//	name.Type, name.Value, instance, v.Name)
				resultLine += fmt.Sprintf("%s;%s;%s;%s",
					name.Type, name.Value, instance, v.Name)

				for _, function := range functions {
					result, err := applyFunction(values, function)
					if err != nil {
						fmt.Fprint(os.Stderr, "Error applying function:", err, "\n")
						continue
					}
					resultLine += fmt.Sprintf(";%.2f", result)
				}
				resultLine += fmt.Sprintf(";%s;", units)

			}
		}
		// delete last semicolon character
		resultLine = resultLine[:len(resultLine)-1]

		results = append(results, resultLine)

	}
	// print title and then results
	fmt.Println(title)
	for _, result := range results {
		fmt.Println(result)
	}
	return nil
}

func GetHostsStatus(ctx context.Context, err error, v *view.ContainerView, entityToQuery string) error {
	var hss []mo.HostSystem

	err = v.Retrieve(ctx, []string{entityToQuery}, []string{"summary"}, &hss)

	if err != nil {
		return err
	}
	hostFound := false

	fmt.Fprint(os.Stdout, "host;uptimeSec;overallStatus;connectionState;inMaintenanceMode;powerState;standbyMode;bootTime;proxyStatus\n")
	for _, hs := range hss {
		if *entityNameFlag != "all" && hs.Summary.Config.Name != *entityNameFlag {
			continue
		}

		fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
			hs.Summary.Config.Name,
			hs.Summary.QuickStats.Uptime,
			hs.Summary.OverallStatus,
			hs.Summary.Runtime.ConnectionState,
			hs.Summary.Runtime.InMaintenanceMode,
			hs.Summary.Runtime.PowerState,
			hs.Summary.Runtime.StandbyMode,
			hs.Summary.Runtime.BootTime.Format("2006-01-02 15:04:05"),
			"OK")
		//
		hostFound = true
	}
	if !hostFound {
		fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
			*entityNameFlag, 0, "NA", "NA", false, "NA", "NA", "NA", "HOST_NOT_FOUND")
		fmt.Fprintf(os.Stderr, "\nHost %s not found\n", *entityNameFlag)
		os.Exit(0)
	}
	return nil
}

// Function to apply the specified operation
func applyFunction(values []float64, function string) (float64, error) {
	switch function {
	case "avg":
		sum := 0.0
		for _, value := range values {
			sum += value
		}
		return sum / float64(len(values)), nil
	case "min":
		min := values[0]
		for _, value := range values {
			if value < min {
				min = value
			}
		}
		return min, nil
	case "max":
		max := values[0]
		for _, value := range values {
			if value > max {
				max = value
			}
		}
		return max, nil
	case "last":
		return values[len(values)-1], nil
	default:
		return 0, fmt.Errorf("unknown function: %s", function)
	}
}

func parseCSV(csv string) ([]float64, error) {
	parts := strings.Split(csv, ",")
	values := make([]float64, 0, len(parts))
	for _, part := range parts {
		value, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
