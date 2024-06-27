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
	"time"
)

var urlFlag = flag.String("url", "", "Required. Usage: -url <https://username:password@host/sdk> (domain users can be set as username@domain)")
var insecureFlag = flag.Bool("insecure", false, "Required. Usage: -insecure")
var entityFlag = flag.String("entity", "host", "Optional. Usage: -entity <host|vm|resourcepool>")
var contextFlag = flag.String("context", "status", "Optional. Usage: -context <status|config|metrics>")
var entityNameFlag = flag.String("entityName", "all", "Optional. Usage: -entityname <host, vm or resource name>")
var timeoutFlag = flag.Duration("timeout", 10*time.Second, "Optional. Usage: -timeout <timeout in duration Ex.: 10s (ms,h,m can be used as well)>")
var intervalFlag = flag.Int("i", 20, "Interval ID")
var metricFlag = flag.String("metric", "cpu.usage.average", "Optional. Usage: -metric <cpu.usage.average>")
var maxSamplesFlag = flag.Int("maxSamples", 1, "Optional. Usage: -maxSamples <number of samples>")

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
	if *entityFlag != "host" {
		fmt.Fprint(os.Stdout, "Option not implemented, set entity to host.\n")
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
		if *entityFlag == "host" && *contextFlag == "status" {
			v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
			if err != nil {
				return err
			}
			defer v.Destroy(ctx)
			return GetHostsStatus(ctx, err, v)
		} else if *entityFlag == "host" && *contextFlag == "metrics" {
			if *metricFlag == "" || *intervalFlag == 0 {
				fmt.Fprint(os.Stdout, "You must specify metrics and interval.\n")
				flag.Usage()
				os.Exit(0)
			}
			v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
			if err != nil {
				return err
			}
			defer v.Destroy(ctx)
			return GetHostMetrics(ctx, err, v)
		}
		fmt.Fprint(os.Stdout, "Option not implemented. Set host staus or host metrics.\n")
		flag.Usage()
		os.Exit(0)
		return nil
	})
}

func GetHostMetrics(ctx context.Context, err error, v *view.ContainerView) error {

	vmsRefs, err := v.Find(ctx, []string{"HostSystem"}, nil)
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

	var names []string
	for name := range counters {
		if *metricFlag != "all" && name != *metricFlag {
			continue
		}
		names = append(names, name)
	}

	// Create PerfQuerySpec
	spec := types.PerfQuerySpec{
		MaxSample:  int32(*maxSamplesFlag),
		MetricId:   []types.PerfMetricId{{Instance: "*"}},
		IntervalId: int32(*intervalFlag),
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, names, vmsRefs)
	if err != nil {
		return err
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return err
	}

	// Read result
	for _, metric := range result {
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
				fmt.Printf("%s\t%s\t%s\t%s\t%s\n",
					name, instance, v.Name, v.ValueCSV(), units)
			}
		}
	}
	return nil
}

func GetHostsStatus(ctx context.Context, err error, v *view.ContainerView) error {
	var hss []mo.HostSystem

	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)

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
