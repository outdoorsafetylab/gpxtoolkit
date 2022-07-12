package controller

import (
	"encoding/json"
	"net/http"
)

type VersionController struct {
	Commit string
	Tag    string
}

func (c *VersionController) Handler(w http.ResponseWriter, r *http.Request) {
	version := &struct {
		Commit string `json:"commit"`
		Tag    string `json:"tag"`
	}{
		Commit: c.Commit,
		Tag:    c.Tag,
	}
	enc := json.NewEncoder(w)
	err := enc.Encode(version)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
