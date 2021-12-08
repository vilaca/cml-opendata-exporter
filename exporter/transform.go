package main

var (
	sensorTypes = map[string]string{
		"ME": "weather",
		"QA": "air quality",
		"RU": "noise",
		"CT": "vehicle counter",
	}
	sensorDescriptions = map[string]map[string]string{
		"C6H6": {"description": "benzene"},
		"00CO": {"description": "carbon monoxide"},
		"00HR": {"description": "relative humidity"},
		"LAEQ": {"description": "equivalent continuous sound level"},
		"0NO2": {"description": "nitrogen dioxide"},
		"00NO": {"description": "nitrogen oxide"}, // undocumented
		"00O3": {"description": "ozone"},
		"00PA": {"description": "atmospheric pressure"},
		"PM10": {"description": "particles with a diameter of less than 10µm"},
		"PM25": {"description": "particles with a diameter of less than 2.5µm"},
		"0SO2": {"description": "sulfur dioxide"},
		"TEMP": {"description": "temperature"},
		"0VTH": {"description": "hourly traffic volume"},
		"00UV": {"description": "ultraviolet"},
		"00VD": {"description": "wind direction"},
		"00VI": {"description": "wind intensity"},
	}
)

func decodeSensorName(name string) (string, string, string) {
	return name[:2], name[2:6], name[:6]
}

func labels(measurement sensorReading) map[string]string {
	sensorType, description, prefix := decodeSensorName(measurement.ID)
	labels := make(map[string]string)
	labels["type"] = sensorTypes[sensorType]
	labels["id"] = measurement.ID
	labels["prefix"] = prefix
	labels["address"] = measurement.Address
	labels["unit"] = measurement.Unit
	labels["date"] = measurement.Date + " " + measurement.DateStandard
	for name, value := range sensorDescriptions[description] {
		labels[name] = value
	}
	return labels
}
