package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/vmware/govmomi/session/cache"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"os"
	"strings"
	"time"
)

var (
	urlFlag        string
	insecureFlag   bool
	entityFlag     string
	entityNameFlag string
	hostedByFlag   string
	timeoutFlag    time.Duration
	intervalFlag   int
	metricsFlag    string
	functionsFlag  string
	maxSamplesFlag int
	instanceFlag   string
)

// NewClient creates a vim25.Client for use in the examples
func NewClient(ctx context.Context) (*vim25.Client, error) {
	u, err := soap.ParseURL(urlFlag)
	if err != nil {
		return nil, err
	}

	s := &cache.Session{
		URL:      u,
		Insecure: insecureFlag,
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
	var err error
	var c *vim25.Client

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeoutFlag)
	defer cancel()

	if urlFlag == "simulator" {
		err = simulator.VPX().Run(f)
		os.Exit(0)
	} else {
		if urlFlag == "" {
			fmt.Fprint(os.Stdout, "You must specify an url.\n")
			os.Exit(1)
		}
		c, err = NewClient(ctx)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Fprintln(os.Stderr, "TIMEOUT")
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "UNABLE_TO_CONNECT")
		os.Exit(1)
	}

	err = f(ctx, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "itoss-vsphere",
		Short:   "Itoss CLI to get VMware vSphere health status, stats and configuration.\nRelies on govmomi client to get VMware vSphere information.",
		Version: "1.0.020",
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	/*

			Global flags available for all commands:
			urlFlag: URL of the vCenter server,
			insecureFlag: Insecure flag to skip SSL verification,
			timeoutFlag: Timeout for the connection to the vCenter server,


			Command flags:

			statusCmd:
				hostFlag: Host entity name flag,
				vmFlag: VM entity name flag,
				clusterFlag: Cluster entity name flag,
				datastoreFlag: Datastore entity name flag,
				resourcePoolFlag: Resource pool entity name flag

			statsCmd:
				metricsFlag: Metrics to query,
				functionsFlag: Functions to query,
				maxSamplesFlag: Maximum number of samples to query,
				instanceFlag: Instance name to query
				hostFlag: Host entity name flag,
				vmFlag: VM entity name flag
				clusterFlag: Cluster entity name flag,
				datastoreFlag: Datastore entity name flag,
				resourcePoolFlag: Resource pool entity name flag

		`	sensorsCmd:
				hostFlag: Host entity name flag,

			configCmd:
				hostFlag: Host entity name flag,
				vmFlag: VM entity name flag,
				clusterFlag: Cluster entity name flag,
				datastoreFlag: Datastore entity name flag,
				resourcePoolFlag: Resource pool entity name flag

	*/
	rootCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "Required. Usage: -u or --url <https://username:password@host/sdk> (domain users can be set as username@domain)")
	rootCmd.PersistentFlags().BoolVarP(&insecureFlag, "insecure", "i", false, "Required. Usage: -i or --insecure")
	rootCmd.PersistentFlags().DurationVarP(&timeoutFlag, "timeout", "t", 10*time.Second, "Optional. Usage: -t or --timeout <timeout in duration Ex.: 10s (ms,h,m can be used as well)>")

	rootCmd.PersistentFlags().StringVarP(&entityFlag, "entity", "e", "host", "Optional. Usage: -e or --entity <host|vm|cluster|datastore|resourcePool>")
	rootCmd.PersistentFlags().StringVarP(&entityNameFlag, "entityName", "n", "*", "Optional. Usage: -n or --entityName <host name| vm name |datastore name| cluster name| resourcePool name>")

	// Define additional flag for datastore entity
	rootCmd.PersistentFlags().StringVarP(&hostedByFlag, "hostedBy", "b", "*", "Optional. Usage: --hostedBy <host name> (for datastore entity only)")

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get the status of specified entities",
		Run: func(cmd *cobra.Command, args []string) {
			Run(func(ctx context.Context, c *vim25.Client) error {
				switch entityFlag {
				case "host":
					return GetHostsStatus(ctx, c)
				case "vm":
					return GetVMStatus(ctx, c)
				case "cluster":
					return GetClusterStatus(ctx, c)
				case "datastore":
					return GetDatastoreStatus(ctx, c)
				default:
					fmt.Fprint(os.Stdout, "Option not implemented.\n")
					os.Exit(1)
				}
				return nil
			})
		},
	}

	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Get the stats of specified entities",
		Run: func(cmd *cobra.Command, args []string) {
			if metricsFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify metrics to query.\n")
				os.Exit(1)
			}

			metrics := strings.Split(metricsFlag, ",")
			if len(metrics) > 1 && instanceFlag != "" {
				fmt.Fprint(os.Stdout, "You must specify only one metric when using instance.\n")
				os.Exit(1)
			}

			var functions []string
			if functionsFlag != "last" {
				functions = strings.Split(functionsFlag, ",")
				for _, f := range functions {
					if !strings.Contains("last,min,max,avg", f) {
						fmt.Fprint(os.Stdout, "You must specify a valid function (avg,min,max,last).\n")
						os.Exit(1)
					}
				}
			} else {
				functions = []string{"last"}
			}

			Run(func(ctx context.Context, c *vim25.Client) error {
				switch entityFlag {
				case "host":
					return GetHostStats(ctx, c, functions)
				case "vm":
					return GetVMStats(ctx, c, functions)
				default:
					fmt.Fprint(os.Stdout, "Option not implemented.\n")
					os.Exit(1)
				}
				return nil
			})
		},
	}

	statsCmd.Flags().StringVarP(&metricsFlag, "metrics", "m", "cpu.usage.average", "For context stats only. Optional. Usage: -m or --metrics <cpu.usage.average,mem.usage.average>")
	statsCmd.Flags().StringVarP(&functionsFlag, "functions", "f", "last", "For context stats only. Optional. Usage: -f or --functions <min,max,avg,last>")
	statsCmd.Flags().IntVarP(&maxSamplesFlag, "maxSamples", "s", 1, "For context stats only. Optional. Usage: -s or --maxSamples <number of samples>")
	statsCmd.Flags().StringVarP(&instanceFlag, "instance", "I", "", "For context stats only. Optional. Usage: -I or --instance <instance name> (default is -)")

	sensorsCmd := &cobra.Command{
		Use:   "sensors",
		Short: "Get sensor information for hosts",
		Run: func(cmd *cobra.Command, args []string) {
			if entityFlag == "host" {
				Run(func(ctx context.Context, c *vim25.Client) error {
					return GetHostsSensors(ctx, c)
				})
			} else {
				fmt.Fprint(os.Stdout, "Option not implemented.\n")
				os.Exit(1)
			}
		},
	}

	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Get configuration details of specified entities",
		Run: func(cmd *cobra.Command, args []string) {
			Run(func(ctx context.Context, c *vim25.Client) error {
				switch entityFlag {
				case "host":
					return GetHostsConfig(ctx, c)
				case "vm":
					return GetVMConfig(ctx, c)
				default:
					fmt.Fprint(os.Stdout, "Option not implemented.\n")
					os.Exit(1)
				}
				return nil
			})
		},
	}

	rootCmd.AddCommand(statusCmd, statsCmd, sensorsCmd, configCmd)

	rootCmd.Execute()
}
