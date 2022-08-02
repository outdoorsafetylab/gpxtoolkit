package controller

import (
	"net/http"

	"github.com/gorilla/mux"
)

func boolVar(r *http.Request, name string, preset bool) bool {
	str, exist := mux.Vars(r)[name]
	if !exist {
		var strs []string
		strs, exist = r.URL.Query()[name]
		if exist {
			str = strs[0]
		}
	}
	if !exist {
		return preset
	}
	if str == "" {
		return true
	}
	return str == "true"
}

// func stringVar(r *http.Request, name, preset string) string {
// 	str := mux.Vars(r)[name]
// 	if str == "" {
// 		str = r.URL.Query().Get(name)
// 	}
// 	if str == "" {
// 		return preset
// 	}
// 	return str
// }

// func intVar(r *http.Request, name string, preset int) int {
// 	str := mux.Vars(r)[name]
// 	if str == "" {
// 		str = r.URL.Query().Get(name)
// 	}
// 	if str == "" {
// 		return preset
// 	}
// 	val, err := strconv.ParseInt(str, 10, 64)
// 	if err != nil {
// 		return preset
// 	}
// 	return int(val)
// }
