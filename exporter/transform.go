package main

var (
	sensorTypes = map[string]string{
		"ME": "weather",
		"QA": "air quality",
		"RU": "noise",
		"CT": "vehicle counter",
	}
	sensorDescriptions = map[string]string{
		"C6H6": "benzene",
		"00CO": "carbon monoxide",
		"00HR": "relative humidity",
		"LAEQ": "equivalent continuous sound level",
		"0NO2": "nitrogen dioxide",
		//"00NO": nitrogen oxide???, // undocumented
		"00O3": "ozone",
		"00PA": "atmospheric pressure",
		//"00PP precipitation??undocumented
		"PM10": "particles with a diameter of less than 10µm",
		"PM25": "particles with a diameter of less than 2.5µm",
		"0SO2": "sulfur dioxide",
		"TEMP": "temperature",
		"0VTH": "hourly traffic volume",
		"00UV": "ultraviolet",
		"00VD": "wind direction",
		"00VI": "wind intensity",
	}
)

func decodeSensorName(name string) (string, string, string) {
	return sensorTypes[name[:2]], sensorDescriptions[name[2:6]], name[:6]
}

func labels(measurement sensorReading) map[string]string {
	sensorType, description, prefix := decodeSensorName(measurement.ID)
	if sensorType == "" {
		sensorType = "Undocumented"
	}
	if description == "" {
		description = "Undocumented"
	}
	labels := make(map[string]string)
	labels["type"] = sensorType
	labels["description"] = description
	labels["prefix"] = prefix
	labels["id"] = measurement.ID
	labels["address"] = measurement.Address
	labels["unit"] = measurement.Unit
	labels["date"] = measurement.Date + " " + measurement.DateStandard
	return labels
}
