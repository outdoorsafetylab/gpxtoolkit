package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"gpxtoolkit/log"
)

type responseDumper struct {
	w http.ResponseWriter
	b bytes.Buffer
	s int
}

func (d *responseDumper) Header() http.Header {
	return d.w.Header()
}

func (d *responseDumper) Write(data []byte) (int, error) {
	d.b.Write(data)
	return d.w.Write(data)
}

func (d *responseDumper) WriteHeader(statusCode int) {
	if d.s == 0 {
		d.w.WriteHeader(statusCode)
		d.s = statusCode
	} else {
		log.Warningf("Attempt to write header again: %d", statusCode)
		debug.PrintStack()
	}
}

func Dump(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("Handling: %s %s", r.Method, r.RequestURI)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %s", err.Error()), 500)
			return
		}
		if data != nil {
			r.Body = io.NopCloser(bytes.NewBuffer(data))
		}
		dumper := &responseDumper{w: w}
		handler.ServeHTTP(dumper, r)
		err = dump(r, data, dumper)
		if err != nil {
			log.Errorf("Failed to dump: %s", err.Error())
		}
	})
}

type request struct {
	Method  string
	URI     string
	Proto   string
	Host    string
	Headers http.Header
	Body    interface{}
}

type response struct {
	Code    int
	Headers http.Header
	Body    interface{}
}

func dump(r *http.Request, data []byte, d *responseDumper) error {
	out := &struct {
		Request  *request
		Response *response
	}{
		Request: &request{
			Method:  r.Method,
			URI:     r.RequestURI,
			Proto:   r.Proto,
			Headers: r.Header,
		},
		Response: &response{
			Code: d.s,
		},
	}
	if len(data) > 0 {
		ctype := r.Header.Get("Content-Type")
		if strings.HasPrefix(ctype, "application/json") {
			out.Request.Body = json.RawMessage(data)
		} else if strings.HasPrefix(ctype, "text/") {
			out.Request.Body = string(data)
		}
	}
	if out.Response.Code == 0 {
		out.Response.Code = 200
	}
	out.Response.Headers = d.Header()
	data = d.b.Bytes()
	if len(data) > 0 {
		ctype := d.Header().Get("Content-Type")
		if strings.HasPrefix(ctype, "application/json") {
			out.Response.Body = json.RawMessage(data)
		} else if strings.HasPrefix(ctype, "text/") {
			out.Request.Body = string(data)
		}
	}
	data, err := json.Marshal(out)
	if err != nil {
		log.Errorf("Failed to marshal dump: %s", err.Error())
		return err
	}
	log.Debugf("%s", string(data))
	return nil
}
