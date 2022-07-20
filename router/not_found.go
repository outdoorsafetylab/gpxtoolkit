package router

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
)

type notFoundHandler struct {
	webroot string
}

func (h *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// w.Header().Add("Cache-Control", "no-cache")
	// w.Header().Add("Cache-Control", "no-store")
	// w.Header().Set("Pragma", "no-cache")
	path := fmt.Sprintf("%s%s", h.webroot, r.URL.Path)
	_, err := os.Stat(path)
	if err == nil {
		sum := sha1sum(path)
		if sum != "" {
			match := r.Header.Get("If-None-Match")
			if match == sum {
				w.WriteHeader(304)
				return
			} else {
				w.Header().Set("ETag", sum)
			}
		}
		http.FileServer(http.Dir(h.webroot)).ServeHTTP(w, r)
	} else {
		http.ServeFile(w, r, fmt.Sprintf("%s/index.html", h.webroot))
		// var data []byte
		// f, err := os.Open(fmt.Sprintf("%s/index.html", h.webroot))
		// if err != nil {
		// 	log.Printf("Failed to open index page: %s", err.Error())
		// 	w.WriteHeader(500)
		// 	return
		// }
		// data, err = ioutil.ReadAll(f)
		// if err != nil {
		// 	log.Printf("Failed to read index page: %s", err.Error())
		// 	w.WriteHeader(500)
		// 	return
		// }
		// w.Header().Set("Content-Type", "text/html")
		// _, err = w.Write(data)
		// if err != nil {
		// 	log.Printf("Failed to write index page: %s", err.Error())
		// 	w.WriteHeader(500)
		// 	return
		// }
	}
}

func sha1sum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
