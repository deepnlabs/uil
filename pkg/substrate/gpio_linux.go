package substrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type LinuxHardwareDriver struct {
	thermalZonePath string
	maxTempCelsius  float64
}

// NewLinuxHardwareDriver initializes the driver targeting the primary CPU thermal zone.
func NewLinuxHardwareDriver(maxTemp float64) *LinuxHardwareDriver {
	if maxTemp <= 0 {
		maxTemp = 80.0 // Default 80°C threshold
	}
	return &LinuxHardwareDriver{
		thermalZonePath: "/sys/class/thermal/thermal_zone0/temp",
		maxTempCelsius:  maxTemp,
	}
}

func (d *LinuxHardwareDriver) ID() string   { return "linux-sysfs-v1" }
func (d *LinuxHardwareDriver) Name() string { return "Linux Sysfs Thermal & Interlock Monitor" }

// ReadCurrentTemperature scans ALL thermal zones and returns the hottest valid one.
// This works on Intel, AMD Ryzen AI, ARM SBCs, Jetson, and heterogeneous hardware.
func (d *LinuxHardwareDriver) ReadCurrentTemperature() (float64, error) {
    zones, err := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
    if err != nil || len(zones) == 0 {
        return 20.0, fmt.Errorf("no thermal zones found")
    }

    maxTemp := 0.0

    for _, zone := range zones {
        data, err := os.ReadFile(zone)
        if err != nil {
            continue
        }

        rawStr := strings.TrimSpace(string(data))
        rawMilli, err := strconv.Atoi(rawStr)
        if err != nil {
            continue
        }

        tempC := float64(rawMilli) / 1000.0

        // Filter out bogus zones
        if tempC < 25.0 { // dummy zones (Intel/AMD zone0)
            continue
        }
        if tempC > 120.0 { // invalid readings
            continue
        }

        if tempC > maxTemp {
            maxTemp = tempC
        }
    }

    if maxTemp == 0.0 {
        return 20.0, fmt.Errorf("no valid thermal zones")
    }

    return maxTemp, nil
}

// Initialize now simply verifies that at least one thermal zone exists.
func (d *LinuxHardwareDriver) Initialize(ctx context.Context, config map[string]any) error {
    zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
    if len(zones) == 0 {
        return fmt.Errorf("no thermal zones found on system")
    }
    return nil
}


func (d *LinuxHardwareDriver) EvaluateInterlock(metrics map[string]float64) (bool, error) {
	currentTemp, err := d.ReadCurrentTemperature()
	if err != nil {
		return false, err
	}

	// Update supplied metric map with live system hardware data
	if metrics != nil {
		metrics["cpu_temp_celsius"] = currentTemp
	}

	// Trigger physical safety interlock breach if temperature exceeds configured limit
	if currentTemp >= d.maxTempCelsius {
		return true, nil
	}

	return false, nil
}

func (d *LinuxHardwareDriver) Shutdown(ctx context.Context) error {
	return nil
}
