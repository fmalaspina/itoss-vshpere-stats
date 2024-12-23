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
	urlFlag      string
	insecureFlag bool
	timeoutFlag  time.Duration

	hostFlag         string
	vmFlag           string
	clusterFlag      string
	datastoreFlag    string
	mountedOnFlag    string
	resourcePoolFlag string

	metricsFlag    string
	functionsFlag  string
	maxSamplesFlag int
	instanceFlag   string
	//versionFlag    bool
	intervalFlag    int
	listMetricsFlag bool
	statusFlag      bool = false
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
			fmt.Fprint(os.Stdout, "You must specify an url. Use -u or --url flag.\n")
			os.Exit(1)
		}
		c, err = NewClient(ctx)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		message := "TIMEOUT"
		if statusFlag {
			switch {
			case hostFlag != "":
				showHostStatusError(message)
			case vmFlag != "":
				showVMStatusError(message)
			case clusterFlag != "":
				showClusterStatusError(message)
			case datastoreFlag != "":
				showDatastoreStatusError(message)
			case resourcePoolFlag != "":
				showResourcePoolStatusError(message)
			}
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, message)
			os.Exit(1)
		}

	} else if err != nil {
		message := "UNABLE_TO_CONNECT"
		if statusFlag {
			switch {
			case hostFlag != "":
				showHostStatusError(message)
			case vmFlag != "":
				showVMStatusError(message)
			case clusterFlag != "":
				showClusterStatusError(message)
			case datastoreFlag != "":
				showDatastoreStatusError(message)
			case resourcePoolFlag != "":
				showResourcePoolStatusError(message)
			}
			os.Exit(0)
		} else {
			fmt.Fprintln(os.Stderr, message)
			os.Exit(1)
		}
	}

	err = f(ctx, c)
	if err != nil {
		if statusFlag {
			switch {
			case hostFlag != "":
				showHostStatusError(err.Error())
			case vmFlag != "":
				showVMStatusError(err.Error())
			case clusterFlag != "":
				showClusterStatusError(err.Error())
			case datastoreFlag != "":
				showDatastoreStatusError(err.Error())
			case resourcePoolFlag != "":
				showResourcePoolStatusError(err.Error())
			}
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "\nError: %s\n", err)
			os.Exit(1)
		}
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:     "itoss-vsphere",
		Short:   "Itoss CLI to get VMware vSphere health status, stats and configuration.\nRelies on govmomi client to get VMware vSphere information.",
		Version: "1.0.029",
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Global flags available for all commands
	rootCmd.PersistentFlags().StringVarP(&urlFlag, "url", "u", "", "(Required) Usage: -u or --url <https://username:password@host/sdk> (domain users can be set as username@domain)")
	rootCmd.PersistentFlags().BoolVarP(&insecureFlag, "insecure", "i", false, "Usage: -i or --insecure")
	rootCmd.PersistentFlags().DurationVarP(&timeoutFlag, "timeout", "T", 10*time.Second, "Usage: -T or --timeout <timeout in duration Ex.: 10s (ms,h,m can be used as well)>")
	rootCmd.PersistentFlags().BoolP("help", "?", false, "Display help information")

	// Status command with specific flags
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get the status of specified entities",
		Run: func(cmd *cobra.Command, args []string) {

			statusFlag = true
			if datastoreFlag != "" && mountedOnFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify host when using datastore. Use -o or --mountedOn flag.\n")
				os.Exit(1)
			}

			Run(func(ctx context.Context, c *vim25.Client) error {
				switch {
				case hostFlag != "":
					return GetHostsStatus(ctx, c)
				case vmFlag != "":
					return GetVMStatus(ctx, c)
				case clusterFlag != "":
					return GetClusterStatus(ctx, c)
				case datastoreFlag != "":
					return GetDatastoreStatus(ctx, c)
				case resourcePoolFlag != "":
					return GetResourcePoolStatus(ctx, c)
				default:
					fmt.Fprint(os.Stdout, "Option not implemented.\n")
					os.Exit(1)
				}
				return nil
			})
		},
	}
	statusCmd.Flags().StringVarP(&hostFlag, "host", "h", "", "Usage: -h or --host <host name>")
	statusCmd.Flags().StringVarP(&vmFlag, "vm", "v", "", "Usage: -v or --vm <vm name>")
	statusCmd.Flags().StringVarP(&clusterFlag, "cluster", "c", "", "Usage: -c or --cluster <cluster name>")
	statusCmd.Flags().StringVarP(&datastoreFlag, "datastore", "d", "", "Usage: -d or --datastore <datastore name>")
	statusCmd.Flags().StringVarP(&mountedOnFlag, "mountedOn", "o", "", "Usage: -o or --mountedOn <host name> (only for Datastore)")
	statusCmd.Flags().StringVarP(&resourcePoolFlag, "resourcePool", "r", "", "Usage: -r or --resourcePool <resource pool name>")

	// Stats command with specific flags
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Get the stats of specified entities",
		Run: func(cmd *cobra.Command, args []string) {
			if listMetricsFlag {
				Run(func(ctx context.Context, c *vim25.Client) error {
					return ListMetrics(ctx, c)
				})
				os.Exit(0)
			}

			if metricsFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify metrics to query. Use -m or --metric flag.\n")
				os.Exit(1)
			}

			if hostFlag == "" && vmFlag == "" && clusterFlag == "" && datastoreFlag == "" && resourcePoolFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify host, vm, cluster, datastore or resourcePool flags.\n")
				os.Exit(1)
			}

			if datastoreFlag != "" && mountedOnFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify host when using datastore. Use -o or --mountedOn flag.\n")
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
				switch {
				case hostFlag != "":
					return GetHostStats(ctx, c, functions)
				case vmFlag != "":
					return GetVMStats(ctx, c, functions)
				case clusterFlag != "":
					return GetClusterStats(ctx, c, functions)
				case datastoreFlag != "":
					return GetDatastoreStats(ctx, c, functions)
				case resourcePoolFlag != "":
					return GetResourcePoolStats(ctx, c, functions)
				default:
					fmt.Fprint(os.Stdout, "Option not implemented.\n")
					cmd.Help()
					os.Exit(1)
				}
				return nil
			})
		},
	}
	statsCmd.Flags().StringVarP(&metricsFlag, "metrics", "m", "", "Usage: -m or --metrics <cpu.usage.average,mem.usage.average>")
	statsCmd.Flags().StringVarP(&functionsFlag, "functions", "f", "last", "Usage: -f or --functions <min,max,avg,last>")
	statsCmd.Flags().IntVarP(&maxSamplesFlag, "maxSamples", "s", 1, "Usage: -s or --maxSamples <number of samples>")
	statsCmd.Flags().IntVarP(&intervalFlag, "interval", "t", 20, "Usage: -t <interval seconds>")
	statsCmd.Flags().StringVarP(&instanceFlag, "instance", "I", "", "Usage: -I or --instance <instance name>")
	statsCmd.Flags().StringVarP(&hostFlag, "host", "h", "", "Usage: --host <host name>")
	statsCmd.Flags().StringVarP(&vmFlag, "vm", "v", "", "Usage: -v or --vm <vm name>")
	statsCmd.Flags().StringVarP(&clusterFlag, "cluster", "c", "", "Usage: -c or --cluster <cluster name>")
	statsCmd.Flags().StringVarP(&datastoreFlag, "datastore", "d", "", "Usage: -d or --datastore <datastore name>")
	statsCmd.Flags().StringVarP(&mountedOnFlag, "mountedOn", "o", "", "Usage: -o or --mountedOn <host name> (only for Datastore)")
	statsCmd.Flags().StringVarP(&resourcePoolFlag, "resourcePool", "r", "", "Usage: -r or --resourcePool <resource pool name>")

	statsCmd.Flags().BoolVarP(&listMetricsFlag, "list", "l", false, "Usage: -l or --list")
	// Sensors command with specific flag
	sensorsCmd := &cobra.Command{
		Use:   "sensors",
		Short: "Get sensor information for hosts",
		Run: func(cmd *cobra.Command, args []string) {
			if hostFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify the --host or -h flag for sensors command.\n")
				os.Exit(1)
			}
			Run(func(ctx context.Context, c *vim25.Client) error {
				return GetHostsSensors(ctx, c)
			})
		},
	}
	sensorsCmd.Flags().StringVarP(&hostFlag, "host", "h", "", "Usage: -h or --host <host name>")

	// Config command with specific flags
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Get configuration details of specified entities",
		Run: func(cmd *cobra.Command, args []string) {
			if datastoreFlag != "" && mountedOnFlag == "" {
				fmt.Fprint(os.Stdout, "You must specify host when using datastore. Use -o or --mountedOn flag.\n")
				os.Exit(1)
			}
			Run(func(ctx context.Context, c *vim25.Client) error {
				switch {

				case hostFlag != "":
					return GetHostsConfig(ctx, c)
				case vmFlag != "":
					return GetVMConfig(ctx, c)
				case clusterFlag != "":
					return GetClusterConfig(ctx, c)
				case datastoreFlag != "":
					return GetDatastoreConfig(ctx, c)
				case resourcePoolFlag != "":
					return GetResourcePoolConfig(ctx, c)
				default:
					fmt.Fprint(os.Stdout, "You must specify host, vm, cluster, datastore or resourcePool flags.\n")
					cmd.Help()
					os.Exit(1)
				}
				return nil
			})

		},
	}
	configCmd.Flags().StringVarP(&hostFlag, "host", "h", "", "Usage: -h or --host <host name>")
	configCmd.Flags().StringVarP(&vmFlag, "vm", "v", "", "Usage: -v or --vm <vm name>")
	configCmd.Flags().StringVarP(&clusterFlag, "cluster", "c", "", "Usage: -c or --cluster <cluster name>")
	configCmd.Flags().StringVarP(&datastoreFlag, "datastore", "d", "", "Usage: -d or --datastore <datastore name>")
	configCmd.Flags().StringVarP(&mountedOnFlag, "mountedOn", "o", "", "Usage: -o or --mountedOn <host name> (only for Datastore)")
	configCmd.Flags().StringVarP(&resourcePoolFlag, "resourcePool", "r", "", "Usage: -r or --resourcePool <resource pool name>")

	rootCmd.AddCommand(statusCmd, statsCmd, sensorsCmd, configCmd)

	rootCmd.Execute()
}
