package controller

import (
	"net/http"

	"github.com/outdoorsafetylab/gpxtoolkit/version"
)

type VersionController struct{}

func (c *VersionController) Get(w http.ResponseWriter, r *http.Request) {
	version := &struct {
		Commit string `json:"commit"`
		Tag    string `json:"tag"`
	}{
		Commit: version.GitHash(),
		Tag:    version.GitTag(),
	}
	writeJSON(w, r, version)
}
