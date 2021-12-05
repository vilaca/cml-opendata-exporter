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

type SensorReading struct {
	Id      string
	Address string
	Value   float64
}

const (
	CmlOpendataUrl        = "http://opendata-cml.qart.pt:8080/lastmeasurements"
	WrongMeasurementValue = -99
	PollingInterval       = 5 * time.Minute
)

var (
	sensorTypes = map[string]string{
		"ME": "weather",
		"QA": "air quality",
		"RU": "noise",
		"CT": "vehicle counter",
	}
	sensorDescriptions = map[string]map[string]string{
		"C6H6": {
			"description": "benzene",
			"unit":        "µg/m3"},
		"00CO": {
			"description": "carbon monoxide",
			"unit":        "µg/m3"},
		"00HR": {
			"description": "relative humidity",
			"unit":        "%"},
		"LAEQ": {
			"description": "equivalent continuous sound level",
			"unit":        "dB(A)"},
		"0NO2": {
			"description": "nitrogen dioxide",
			"unit":        "µg/m3"},
		"00NO": { // undocumented
			"description": "nitrogen oxide",
			"unit":        "µg/m3"},
		"00O3": {
			"description": "ozone",
			"unit":        "µg/m3"},
		"00PA": {
			"description": "atmospheric pressure",
			"unit":        "mbar"},
		"PM10": {
			"description": "particles with a diameter of less than 10µm",
			"unit":        "µg/m3"},
		"PM25": {
			"description": "particles with a diameter of less than 2.5µm",
			"unit":        "µg/m3"},
		"0SO2": {
			"description": "sulfur dioxide",
			"unit":        "µg/m3"},
		"TEMP": {
			"description": "temperature",
			"unit":        "ºC"},
		"0VTH": {
			"description": "hourly traffic volume",
			"unit":        "vehicles"},
		"00UV": {
			"description": "ultraviolet"},
		"00VD": {
			"description": "wind direction",
			"unit":        "º"},
		"00VI": {
			"description": "wind intensity",
			"unit":        "km/h"},
	}
	cache = map[string]prometheus.Gauge{}
	runs  = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_runs",
		Help: "The total number of runs",
	})
	downloadError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_download_error",
		Help: "The total number of download errors",
	})
	badRead = promauto.NewCounter(prometheus.CounterOpts{
		Name: "lx_sensor_bad_reads",
		Help: "The total number of bad reads",
	})
	downloadTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_download_time",
		Help: "Amount of time it took to download metrics in milliseconds",
	})
	executionTime = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "lx_sensor_pooling_execution_time",
		Help: "Amount of time it took to pool and update metrics in milliseconds",
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
func decodeSensorName(name string) (string, string, string) {
	return name[0:2], name[2:6], name[6:]
}
func labels(measurement SensorReading) map[string]string {
	sensorType, description, numericId := decodeSensorName(measurement.Id)
	labels := make(map[string]string)
	labels["type"] = sensorTypes[sensorType]
	labels["id"] = numericId
	labels["key"] = measurement.Id
	labels["address"] = measurement.Address
	for name, value := range sensorDescriptions[description] {
		labels[name] = value
	}
	return labels
}
func newGauge(measurement SensorReading) prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "lx_sensor_measurement",
		ConstLabels: labels(measurement),
		Help:        "Sensor measurement",
	})
}
func updateMetric(measurement SensorReading) {
	key := measurement.Id
	gauge := cache[key]
	if gauge == nil {
		gauge = newGauge(measurement)
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