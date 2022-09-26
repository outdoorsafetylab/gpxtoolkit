package server

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

type weboortHandler struct {
	path string
}

func (h *weboortHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := fmt.Sprintf("%s%s", h.path, r.URL.Path)
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		path = fmt.Sprintf("%s/index.html", h.path)
	}
	sum, err := sha1sum(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	match := r.Header.Get("If-None-Match")
	if match == sum {
		w.WriteHeader(304)
		return
	}
	w.Header().Set("ETag", sum)
	http.ServeFile(w, r, path)
}

var sha1sumCache = map[string]string{}

func sha1sum(path string) (string, error) {
	sum := sha1sumCache[path]
	if sum != "" {
		return sum, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	sum = hex.EncodeToString(h.Sum(nil))
	sha1sumCache[path] = sum
	return sum, nil
}
