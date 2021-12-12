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

type sensorReading struct {
	ID           string
	Address      string
	Avg          string
	Unit         string
	Value        float64
}

const (
	CmlOpenDataURL        = "http://opendata-cml.qart.pt:8080/lastmeasurements"
	WrongMeasurementValue = -99
	PollingInterval       = 5 * time.Minute
)

var (
	cache = map[string]prometheus.Gauge{}

	runs = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_pooling_runs",
		Help: "The total number of runs",
	})
	badRead = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_pooling_measurement_error",
		Help: "The number of bad reads",
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
		Name: "lx_sensor_pooling_execution_time",
		Help: "Amount of time it took to pool and update metrics in milliseconds",
	})
	totalMeasurements = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_pooling_measurement_total",
		Help: "Amount of downloaded measurements",
	})
)

func downloadMeasurements() ([]sensorReading, error) {
	start := time.Now()
	resp, err := http.Get(CmlOpenDataURL)
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
	var measurements []sensorReading
	err = json.Unmarshal(body, &measurements)
	if err != nil {
		return nil, err
	}
	downloadTime.Set(float64(time.Since(start).Milliseconds()))
	return measurements, nil
}

func newGauge(measurement sensorReading, key string) prometheus.Gauge {
	gauge := cache[key]
	if gauge == nil {
		opts := prometheus.GaugeOpts{
			Name:        "lx_sensor_measurement",
			Help:        "Sensor measurement",
			ConstLabels: labels(measurement),
		}
		gauge = prometheus.NewGauge(opts)
		prometheus.MustRegister(gauge)
		cache[key] = gauge
	}
	return gauge
}

func recordMetrics() {
	start := time.Now()
	measurements, err := downloadMeasurements()
	if err != nil {
		downloadError.Inc()
		log.Println(err)
		return
	}
	totalMeasurements.Set(float64(len(measurements)))
	badReads := 0
	for _, measurement := range measurements {
		if measurement.Value == WrongMeasurementValue {
			badReads++
			continue
		}
		updateMetric(measurement)
	}
	badRead.Set(float64(badReads))
	runs.Inc()
	elapsed := time.Since(start)
	executionTime.Set(float64(elapsed.Milliseconds()))
}

func updateMetric(measurement sensorReading) {
	key := measurement.ID
	gauge := newGauge(measurement, key)
	gauge.Set(measurement.Value)
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
