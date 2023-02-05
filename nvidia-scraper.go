package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	memoryTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_memory_total",
			Help: "Total memory for each GPU",
		},
		[]string{"gpu"},
	)
	memoryFree = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_memory_free",
			Help: "Free memory for each GPU",
		},
		[]string{"gpu"},
	)
	memoryUsed = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_memory_used",
			Help: "Used memory for each GPU",
		},
		[]string{"gpu"},
	)
	utilizationGPU = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_utilization_gpu",
			Help: "GPU utilization for each GPU",
		},
		[]string{"gpu"},
	)
	utilizationMemory = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_utilization_memory",
			Help: "Memory utilization for each GPU",
		},
		[]string{"gpu"},
	)
	temperature = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_temperature",
			Help: "Temperature for each GPU",
		},
		[]string{"gpu"},
	)
	powerUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nvidia_power_usage",
			Help: "Power usage for each GPU",
		},
		[]string{"gpu"},
	)
)

func main() {
	// Initialize NVML
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		fmt.Printf("error initializing NVML: %s\n", ret)
		return
	}
	defer nvml.Shutdown()

	// Get the number of GPUs
	deviceCount, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		fmt.Printf("error getting device count: %s\n", ret)
		return
	}

	// Iterate through the GPUs
	for i := int(0); i < deviceCount; i++ {
		// Get handle to GPU
		handle, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			fmt.Printf("error getting device handle for GPU %d: %s\n", i, ret)
			continue
		}

		// Get memory info
		memoryInfo, ret := handle.GetMemoryInfo()
		if ret != nvml.SUCCESS {
			fmt.Printf("error getting memory info for GPU %d: %s\n", i, ret)
		} else {
			memoryTotal.WithLabelValues(strconv.Itoa(i)).Set(float64(memoryInfo.Total / 1024 / 1024))
			memoryFree.WithLabelValues(strconv.Itoa(i)).Set(float64(memoryInfo.Free / 1024 / 1024))
			memoryUsed.WithLabelValues(strconv.Itoa(i)).Set(float64(memoryInfo.Used / 1024 / 1024))
		}

		// Get utilization rates
		utilization, ret := handle.GetUtilizationRates()
		if ret != nvml.SUCCESS {
			fmt.Printf("error getting utilization rates for GPU %d: %s\n", i, ret)
		} else {
			utilizationGPU.WithLabelValues(strconv.Itoa(i)).Set(float64(utilization.Gpu))
			utilizationMemory.WithLabelValues(strconv.Itoa(i)).Set(float64(utilization.Memory))
		}

		// Get temperature
		var tempSensors nvml.TemperatureSensors
		temp, ret := handle.GetTemperature(tempSensors)
		if ret != nvml.SUCCESS {
			fmt.Printf("error getting temperature for GPU %d: %s\n", i, ret)
		} else {
			temperature.WithLabelValues(strconv.Itoa(i)).Set(float64(temp))
		}

		power, ret := handle.GetPowerUsage()
		if ret != nvml.SUCCESS {
			fmt.Printf("error getting power usage for GPU %d: %s\n", i, ret)
		} else {
			powerUsage.WithLabelValues(strconv.Itoa(i)).Set(float64(power))
		}
	}
	// Create HTTP handler for Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	fmt.Println("Starting server at :8080")
	http.ListenAndServe(":8080", nil)

}
