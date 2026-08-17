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

func (d *LinuxHardwareDriver) Initialize(ctx context.Context, config map[string]any) error {
	// Verify thermal zone presence on host system
	if _, err := os.Stat(d.thermalZonePath); os.IsNotExist(err) {
		// Fallback search for alternative thermal zones on ARM/x86 SBCs
		zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
		if len(zones) > 0 {
			d.thermalZonePath = zones[0]
		} else {
			return fmt.Errorf("no valid sysfs thermal zone detected at %s", d.thermalZonePath)
		}
	}
	return nil
}

// ReadCurrentTemperature parses raw millidegrees Celsius from Linux sysfs
func (d *LinuxHardwareDriver) ReadCurrentTemperature() (float64, error) {
	data, err := os.ReadFile(d.thermalZonePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read thermal sensor: %w", err)
	}

	rawStr := strings.TrimSpace(string(data))
	rawMilli, err := strconv.ParseFloat(rawStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid thermal sensor payload: %w", err)
	}

	// /sys/class/thermal outputs values in millidegrees (e.g., 42500 = 42.5°C)
	return rawMilli / 1000.0, nil
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
