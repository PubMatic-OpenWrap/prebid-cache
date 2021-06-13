package kvserver

import "strconv"

func GetInt(value string) int {
	ivalue, _ := strconv.Atoi(value)
	return ivalue
}

func GetFloat64(value string) float64 {
	fvalue, _ := strconv.ParseFloat(value, 64)
	return fvalue
}
