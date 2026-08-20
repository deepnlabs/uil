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

func NewLinuxHardwareDriver(maxTemp float64) *LinuxHardwareDriver {
    if maxTemp <= 0 {
        maxTemp = 80.0
    }
    return &LinuxHardwareDriver{
        thermalZonePath: "/sys/class/thermal/thermal_zone0/temp",
        maxTempCelsius:  maxTemp,
    }
}

func (d *LinuxHardwareDriver) ID() string        { return "linux-sysfs-v1" }
func (d *LinuxHardwareDriver) Name() string      { return "Linux Sysfs Thermal & Interlock Monitor" }
func (d *LinuxHardwareDriver) APIVersion() int   { return SubstrateAPIVersion }

// Hardened temperature reader with panic‑recovery and strict validation.
func (d *LinuxHardwareDriver) ReadCurrentTemperature() (float64, error) {
    defer func() {
        if r := recover(); r != nil {
            // Never allow panic to propagate into governance loop
        }
    }()

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

        // Strict sanity filtering
        if tempC < 10.0 { // bogus zones
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

func (d *LinuxHardwareDriver) Initialize(ctx context.Context, config map[string]any) error {
    zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
    if len(zones) == 0 {
        return fmt.Errorf("no thermal zones found on system")
    }
    return nil
}

// Hardened EvaluateInterlock with panic‑recovery and metric validation.
func (d *LinuxHardwareDriver) EvaluateInterlock(metrics map[string]float64) (bool, error) {
    defer func() {
        if r := recover(); r != nil {
            // Never allow plugin or driver panic to crash governance loop
        }
    }()

    currentTemp, err := d.ReadCurrentTemperature()
    if err != nil {
        return false, err
    }

    if metrics != nil {
        metrics["cpu_temp_celsius"] = currentTemp
    }

    if currentTemp >= d.maxTempCelsius {
        return true, nil
    }

    return false, nil
}

func (d *LinuxHardwareDriver) Shutdown(ctx context.Context) error {
    return nil
}
