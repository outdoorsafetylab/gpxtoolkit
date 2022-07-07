package controller

import "net/url"

func getQuery(q url.Values, name, preset string) string {
	val := q.Get(name)
	if val == "" {
		return preset
	}
	return val
}
