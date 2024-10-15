package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"strconv"
	"strings"
	"time"
)

func safeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case *string:
		if v == nil {
			return "NA"
		}
		return *v
	case *time.Time:
		if v == nil {
			return "NA"
		}
		return v.Format("2006-01-02 15:04:05")
	case interface{}:
		if v == nil {
			return "NA"
		}
		return v
	default:
		return value
	}
}

func getVMName(ctx context.Context, v *view.ContainerView, name string) (string, error) {
	var hss []mo.VirtualMachine

	err := v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &hss)
	if err != nil {
		return "", err
	}
	for _, hs := range hss {
		if hs.Summary.Config.Name == name {
			return hs.Summary.Vm.Value, nil
		}
	}
	return "", fmt.Errorf("vm %s not found", name)
}

func parseMap(s string) map[string]string {
	result := make(map[string]string)

	// Split the input string by comma to separate key-value pairs
	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		// Split each pair by the equals sign
		keyValue := strings.SplitN(pair, "=", 2)
		if len(keyValue) == 2 {
			key := keyValue[0]
			value := keyValue[1]
			result[key] = value
		}
	}
	return result
}
func contains(names []string, value string) bool {
	for _, name := range names {
		if name == value {
			return true
		}
	}
	return false
}
func getHostNames(ctx context.Context, c *vim25.Client, name string) ([]string, error) {
	m := view.NewManager(c)
	vHost, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		var emptyHostnames []string
		return emptyHostnames, fmt.Errorf("host %s not found", name)
	}
	defer vHost.Destroy(ctx)
	var hss []mo.HostSystem

	var hostNames []string
	err = vHost.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"summary"}, &hss, property.Match{"name": name})
	if err != nil {
		return hostNames, err
	}

	// create a []string with the hss.Summary.Host.Value

	for _, hs := range hss {
		hostNames = append(hostNames, hs.Summary.Host.Value)
	}

	if len(hostNames) > 0 {
		return hostNames, nil
	}
	return hostNames, fmt.Errorf("host %s not found", name)
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
