package controller

import (
	"encoding/json"
	"gpxtoolkit/version"
	"net/http"
)

type VersionController struct{}

func (c *VersionController) Handler(w http.ResponseWriter, r *http.Request) {
	version := &struct {
		Commit string `json:"commit"`
		Tag    string `json:"tag"`
	}{
		Commit: version.GitHash,
		Tag:    version.GitTag,
	}
	enc := json.NewEncoder(w)
	err := enc.Encode(version)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
