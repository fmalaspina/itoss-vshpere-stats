package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"os"
	"strings"
	"time"
)

var urlFlag = flag.String("url", "", "Required. Usage: -url <https://username:password@host/sdk> (domain users can be set as username@domain)")
var insecureFlag = flag.Bool("insecure", false, "Required. Usage: -insecure")
var entityFlag = flag.String("entity", "host", "Optional. Usage: -entity <host|vm>")
var contextFlag = flag.String("context", "status", "Optional. Usage: -context <status|datastore|stats>")
var entityNameFlag = flag.String("entityName", "all", "Optional. Usage: -entityName <host name| vm name")
var timeoutFlag = flag.Duration("timeout", 10*time.Second, "Optional. Usage: -timeout <timeout in duration Ex.: 10s (ms,h,m can be used as well)>")
var intervalFlag = flag.Int("i", 20, "Optional. Usage: -i <interval id>")
var metricsFlag = flag.String("metrics", "cpu.usage.average", "For context stats only. Optional. Usage: -metrics <cpu.usage.average,mem.usage.average>")
var functionsFlag = flag.String("functions", "last", "For context stats only. Optional. Usage: -functions <min,max,avg,last>")
var maxSamplesFlag = flag.Int("maxSamples", 1, "For context stats only. Optional. Usage: -maxSamples <number of samples>")
var instanceFlag = flag.String("instance", "", "For context stats only. Optional. Usage: -instance <instance name> (default is -)")
var versionFlag = flag.Bool("version", false, "Optional. Usage: -version")

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

	if *contextFlag != "status" && *contextFlag != "stats" && *contextFlag != "datastore" {
		fmt.Fprint(os.Stdout, "Option not implemented, set context to status, stats or datastore.\n")
		flag.Usage()
		os.Exit(1)
	}
	if *entityFlag != "host" && *entityFlag != "vm" {
		fmt.Fprint(os.Stdout, "Option not implemented, set entity to host or vm.\n")
		flag.Usage()
		os.Exit(1)
	}
	if *urlFlag == "simulator" {
		err = simulator.VPX().Run(f)
		os.Exit(0)
	} else {
		if *urlFlag == "" {
			fmt.Fprint(os.Stdout, "You must specify an url.\n")
			flag.Usage()
			os.Exit(1)
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
		os.Exit(0)
	}
	if err != nil && *contextFlag != "status" {
		fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Fprint(os.Stdout, "Version: 1.0.003\n")
		os.Exit(0)
	}
	Run(func(ctx context.Context, c *vim25.Client) error {
		// Create a view of HostSystem objects
		m := view.NewManager(c)
		var entityToQuery = ""

		if *entityFlag == "host" {
			entityToQuery = "HostSystem"
		}

		if *contextFlag == "datastore" {
			entityToQuery = "Datastore"
		}

		if *entityFlag == "vm" {
			entityToQuery = "VirtualMachine"
		}

		v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{entityToQuery}, true)
		if err != nil {
			return err
		}
		defer v.Destroy(ctx)

		if *contextFlag == "status" {
			return GetHostsStatus(ctx, err, v, entityToQuery)
		} else if *contextFlag == "datastore" {
			return GetDatastoreStatus(ctx, err, v)
		} else if *contextFlag == "stats" {
			if *metricsFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify metrics to query.\n")
				flag.Usage()
				os.Exit(1)
			}

			if *metricsFlag != "" {
				metrics := strings.Split(*metricsFlag, ",")
				if len(metrics) > 1 && *instanceFlag != "" {
					fmt.Fprint(os.Stdout, "You must specify only one metric when using instance.\n")
					flag.Usage()
					os.Exit(1)
				}
			}
			var functions []string
			if *functionsFlag != "last" {
				functions = strings.Split(*functionsFlag, ",")

				for _, f := range functions {
					if !strings.Contains("last,min,max,avg", f) {
						fmt.Fprint(os.Stdout, "You must specify a valid function (avg,min,max,last).\n")
						flag.Usage()
						os.Exit(1)
					}
				}
			} else {
				functions = []string{"last"}
			}

			//v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{entityToQuery}, true)
			//if err != nil {
			//	return err
			//}

			return GetHostStats(ctx, err, v, functions, entityToQuery)
		}
		fmt.Fprint(os.Stdout, "Option not implemented. Set host status or host metrics.\n")
		flag.Usage()
		os.Exit(1)
		return nil
	})
}
