package main

type SensorReading struct {
	Id      string
	Address string
	Value   float64
}

var (
	sensorTypes = map[string]string{
		"ME": "weather",
		"QA": "air quality",
		"RU": "noise",
		"CT": "vehicle counter",
	}
	sensorDescriptions = map[string]map[string]string{
		"C6H6": {"description": "benzene", "unit": "µg/m3"},
		"00CO": {"description": "carbon monoxide", "unit": "µg/m3"},
		"00HR": {"description": "relative humidity", "unit": "%"},
		"LAEQ": {"description": "equivalent continuous sound level", "unit": "dB(A)"},
		"0NO2": {"description": "nitrogen dioxide", "unit": "µg/m3"},
		"00NO": {"description": "nitrogen oxide", "unit": "µg/m3"}, // undocumented
		"00O3": {"description": "ozone", "unit": "µg/m3"},
		"00PA": {"description": "atmospheric pressure", "unit": "mbar"},
		"PM10": {"description": "particles with a diameter of less than 10µm", "unit": "µg/m3"},
		"PM25": {"description": "particles with a diameter of less than 2.5µm", "unit": "µg/m3"},
		"0SO2": {"description": "sulfur dioxide", "unit": "µg/m3"},
		"TEMP": {"description": "temperature", "unit": "ºC"},
		"0VTH": {"description": "hourly traffic volume", "unit": "vehicles"},
		"00UV": {"description": "ultraviolet"},
		"00VD": {"description": "wind direction", "unit": "º"},
		"00VI": {"description": "wind intensity", "unit": "km/h"},
	}
)

func decodeSensorName(name string) (string, string, string, string) {
	return name[:2], name[2:6], name[6:], name[:6]
}

func labels(measurement SensorReading) map[string]string {
	sensorType, description, numericId, prefix := decodeSensorName(measurement.Id)
	labels := make(map[string]string)
	labels["type"] = sensorTypes[sensorType]
	labels["id"] = numericId
	labels["prefix"] = prefix
	labels["key"] = measurement.Id
	labels["address"] = measurement.Address
	for name, value := range sensorDescriptions[description] {
		labels[name] = value
	}
	return labels
}
