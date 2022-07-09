package controller

import (
	"log"
	"net/url"
	"strconv"
)

func queryGetString(q url.Values, name, preset string) string {
	val := q.Get(name)
	if val == "" {
		return preset
	}
	return val
}

func queryGetFloat64(q url.Values, name string, preset float64) float64 {
	str := q.Get(name)
	if str == "" {
		log.Printf("Missing '%s'", name)
		return preset
	}
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Printf("Invalid '%s': %s", name, str)
		return preset
	}
	return val
}

func queryGetBool(q url.Values, name string, preset bool) bool {
	str := q.Get(name)
	if str == "" {
		log.Printf("Missing '%s'", name)
		return preset
	}
	return str == "true"
}
