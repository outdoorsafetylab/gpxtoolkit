package main

import (
	"flag"
	"fmt"
	"gpxtoolkit/gpx"
	"gpxtoolkit/milestone"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/gorilla/mux"
)

var progname string
var nameTemplate = `5{{printf "%02d" .Index}}/{{printf "%.1f" .Kilometer}}K`

func help(progname string) {
	fmt.Printf("Usage: %s <gpx>\n", progname)
}

func main() {
	progname = filepath.Base(os.Args[0])
	daemon := false
	port := 8080
	env := os.Getenv("PORT")
	if env != "" {
		val, err := strconv.ParseInt(env, 10, 32)
		if err != nil {
			port = int(val)
		}
	}
	flag.BoolVar(&daemon, "d", daemon, "start a daemon")
	flag.IntVar(&port, "p", port, "HTTP port")
	flag.StringVar(&nameTemplate, "n", nameTemplate, "Template of waypint name")
	flag.Parse()
	var err error
	if daemon {
		err = startDaemon(port)
	} else {
		err = startCommand()
	}
	if err != nil {
		os.Exit(1)
	}
}

func startCommand() error {
	file := flag.Arg(0)
	if file == "" {
		help(progname)
		return fmt.Errorf("Missing argument of input file")
	}
	log, err := gpx.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GPX '%s': %s\n", file, err.Error())
		return err
	}
	tmpl, err := template.New("").Parse(nameTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template: %s\n", err.Error())
		return err
	}
	marker := &milestone.Marker{
		Distance:     100,
		NameTemplate: tmpl,
		Reverse:      false,
	}
	err = marker.Mark(log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to mark GPX: %s\n", err.Error())
		return err
	}
	writer := &gpx.Writer{
		Creator: progname,
		Writer:  os.Stdout,
	}
	err = writer.Write(log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write GPX: %s\n", err.Error())
		return err
	}
	return nil
}

func startDaemon(port int) error {
	r := mux.NewRouter()
	s := r.PathPrefix("/cgi").Subrouter()
	s.HandleFunc("/milestones", milestonesHandler).Methods("POST")
	r.NotFoundHandler = http.FileServer(http.Dir("./webroot"))
	log.Printf("Listening port %d...", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen %d: %s\n", port, err.Error())
	}
	return err
}

func milestonesHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 * 1048576)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	distance, err := strconv.ParseFloat(r.FormValue("distance"), 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid distance: %s", err.Error()), 400)
		return
	}
	f, h, err := r.FormFile("gpx-file")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer f.Close()
	log, err := gpx.Parse(f)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	tmpl, err := template.New("").Parse(r.FormValue("name-template"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse template: %s", err.Error()), 400)
		return
	}
	marker := &milestone.Marker{
		Distance:     distance,
		NameTemplate: tmpl,
	}
	if _, ok := r.Form["reverse"]; ok {
		marker.Reverse = true
	}
	err = marker.Mark(log)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark GPX: %s", err.Error()), 500)
		return
	}
	writer := &gpx.Writer{
		Creator: progname,
		Writer:  w,
	}
	w.Header().Set("Content-Type", "application/gpx+xml")
	extname := filepath.Ext(h.Filename)
	basename := h.Filename[0 : len(h.Filename)-len(extname)]
	filename := fmt.Sprintf("%s%s%s", basename, r.FormValue("filename-suffix"), extname)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename)))
	err = writer.Write(log)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write GPX: %s", err.Error()), 500)
		return
	}
}
