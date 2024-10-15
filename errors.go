package main

import (
	"fmt"
	"os"
)

func showClusterError(errorText string) {
	fmt.Fprint(os.Stdout, "cluster;totalCpu;totalMemory;numCpuCores;numCpuThreads;effectiveCpu;effectiveMemory;numHosts;numEffectiveHosts;overallStatus;proxyStatus\n")

	fmt.Fprintf(os.Stdout, "%s;%d;%d;%d;%d;%d;%d;%d;%d;%s;%s\n",
		clusterFlag, 0, 0, 0, 0, 0, 0, 0, 0, "NA", errorText)
	os.Exit(0)
}

func showDatastoreStatusError(errorText string) {
	fmt.Fprint(os.Stdout, "name;host;internalHostname;type;maintenanceMode;capacity;freeSpace;uncommitted;accessible;proxyStatus; ")
	fmt.Fprintf(os.Stdout, "%s;%s;%s;%s;%s;%s;%s;%v;%s;%s\n",
		"NA", "NA", "NA", "NA", "NA", "NA", "NA", 0, "NA", errorText)
	os.Exit(0)
}

func showVMStatusError(errorText string) {
	fmt.Fprint(os.Stdout, "name;internalName;overallStatus;connectionState;powerState;guestHeartbeatStatus;bootTime;uptimeSeconds;proxyStatus\n")
	fmt.Fprintf(os.Stdout, "%s;%s;%s;%s;%s;%s;%s;%v;%s\n",
		"NA", "NA", "NA", "NA", "NA", "NA", "NA", 0, errorText)
	os.Exit(0)
}

func showHostStatusError(errorText string) {
	fmt.Fprint(os.Stdout, "host;uptimeSec;overallStatus;connectionState;inMaintenanceMode;powerState;standbyMode;bootTime;proxyStatus\n")
	fmt.Fprintf(os.Stdout, "%s;%d;%s;%s;%v;%s;%s;%s;%s\n",
		"NA", 0, "NA", "NA", false, "NA", "NA", "NA", errorText)
	os.Exit(0)
}
