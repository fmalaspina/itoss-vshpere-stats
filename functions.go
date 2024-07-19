package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"strconv"
	"strings"
)

func getEntityName(ctx context.Context, v *view.ContainerView, entityToQuery string, name string) (string, error) {
	var hss []mo.HostSystem

	err := v.Retrieve(ctx, []string{entityToQuery}, []string{"summary"}, &hss)
	if err != nil {
		return "", err
	}
	for _, hs := range hss {
		if hs.Summary.Config.Name == name {
			return hs.Summary.Host.Value, nil
		}
	}
	return "", fmt.Errorf("host %s not found", name)
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
