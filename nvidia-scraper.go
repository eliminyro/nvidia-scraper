package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	gpuFreq = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nvidia_gpu_freq",
		Help: "Current GPU frequency in MHz",
	})

	gpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nvidia_gpu_usage",
		Help: "Current GPU usage in percentage",
	})

	memUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nvidia_mem_used",
		Help: "Current memory usage in bytes",
	})

	memTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nvidia_mem_total",
		Help: "Total memory available in bytes",
	})

	powerDraw = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nvidia_power_draw",
		Help: "Current power draw in Watts",
	})

	gpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nvidia_gpu_temp",
		Help: "Current GPU temperature in Celsius",
	})
)

func main() {
	// Initialize NVML
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatal(ret)
	}
	defer nvml.Shutdown()

	// Register Prometheus metrics
	prometheus.MustRegister(gpuFreq)
	prometheus.MustRegister(gpuUsage)
	prometheus.MustRegister(memUsed)
	prometheus.MustRegister(memTotal)
	prometheus.MustRegister(powerDraw)
	prometheus.MustRegister(gpuTemp)

	// Get the number of GPUs
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		fmt.Println("Error getting device count:", ret)
		return
	}

	// Scrape metrics every 5 seconds

	go func() {
		for {
			for i := int(0); i < deviceCount; i++ {
				device, ret := nvml.DeviceGetHandleByIndex(i)
				if ret != nvml.SUCCESS {
					fmt.Println("Failed to get device:", i, ret)
					continue
				}

				// GPU frequency
				var clockId nvml.ClockId
				var clockType nvml.ClockType
				freq, ret := device.GetClock(clockType, clockId)
				if ret != nvml.SUCCESS {
					fmt.Println("Failed to get GPU clock:", ret)
				} else {
					gpuFreq.Set(float64(freq))
				}

				// GPU usage
				util, ret := device.GetUtilizationRates()
				if ret != nvml.SUCCESS {
					fmt.Println("Failed to get GPU utilization:", ret)
				} else {
					gpuUsage.Set(float64(util.Gpu))
				}

				// Memory usage
				memInfo, ret := device.GetMemoryInfo()
				if ret != nvml.SUCCESS {
					fmt.Println("Failed to get memory info:", ret)
				} else {
					memUsed.Set(float64(memInfo.Used))
					memTotal.Set(float64(memInfo.Total))
				}

				// Power draw
				power, ret := device.GetPowerUsage()
				if ret != nvml.SUCCESS {
					fmt.Println("Failed to get power usage:", ret)
				} else {
					powerDraw.Set(float64(power) / 1000.0)
				}

				// GPU temperature
				var tempSensors nvml.TemperatureSensors
				temp, ret := device.GetTemperature(tempSensors)
				if ret != nvml.SUCCESS {
					fmt.Println("Failed to get GPU temperature:", ret)
				} else {
					gpuTemp.Set(float64(temp))
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// Expose the Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
