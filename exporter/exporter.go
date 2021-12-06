package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	CmlOpendataUrl        = "http://opendata-cml.qart.pt:8080/lastmeasurements"
	WrongMeasurementValue = -99
	PollingInterval       = 5 * time.Minute
)

var (
	cache = map[string]prometheus.Gauge{}

	runs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_pooling_runs",
		Help: "The total number of runs",
	})
	badRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_pooling_measurement_error",
		Help: "The total number of bad reads",
	})
	downloadError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_pooling_download_error",
		Help: "The total number of download errors",
	})
	downloadTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_pooling_download_time",
		Help: "Amount of time it took to download metrics in milliseconds",
	})
	executionTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_execution_time",
		Help: "Amount of time it took to pool and update metrics in milliseconds",
	})
	totalMeasurements = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_pooling_measurement_total",
		Help: "Amount of downloaded measurements",
	})
)

func downloadMeasurements() ([]SensorReading, error) {
	resp, err := http.Get(CmlOpendataUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Print(err)
		}
	}(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var measurements []SensorReading
	err = json.Unmarshal(body, &measurements)
	if err != nil {
		return nil, err
	}
	return measurements, nil
}

func updateMetric(measurement SensorReading) {
	key := measurement.Id
	gauge := cache[key]
	if gauge == nil {
		opts := prometheus.GaugeOpts{
			Name:        "lx_sensor_measurement",
			Help:        "Sensor measurement",
			ConstLabels: labels(measurement),
		}
		gauge = prometheus.NewGauge(opts)
		cache[key] = gauge
		prometheus.MustRegister(gauge)
	}
	gauge.Set(measurement.Value)
}

func recordMetrics() {
	start := time.Now()
	measurements, err := downloadMeasurements()
	elapsed := time.Since(start)
	if err != nil {
		downloadError.Inc()
		log.Println(err)
		return
	}
	downloadTime.Set(float64(elapsed.Milliseconds()))
	totalMeasurements.Set(float64(len(measurements)))
	for _, measurement := range measurements {
		if measurement.Value == WrongMeasurementValue {
			badRead.Inc()
			continue
		}
		updateMetric(measurement)
	}
	runs.Inc()
	elapsed = time.Since(start)
	executionTime.Set(float64(elapsed.Milliseconds()))
}

func main() {
	go func() {
		for {
			recordMetrics()
			time.Sleep(PollingInterval)
		}
	}()
	http.Handle("/", promhttp.Handler())
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println(err)
	}
}
