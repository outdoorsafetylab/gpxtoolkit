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
var command = &Command{
	template: `{{printf "%.1f" .Kilometer}}K`,
	distance: 100,
	reverse:  false,
}
var server = &Server{
	webroot: "./webroot",
	port:    8080,
}

func help(progname string) {
	fmt.Printf("Usage: %s <gpx>\n", progname)
}

func main() {
	progname = filepath.Base(os.Args[0])
	daemon := false
	env := os.Getenv("PORT")
	if env != "" {
		val, err := strconv.ParseInt(env, 10, 32)
		if err != nil {
			server.port = int(val)
		}
	}
	flag.BoolVar(&daemon, "D", daemon, "run as a HTTP server")
	flag.IntVar(&server.port, "p", server.port, "HTTP port")
	flag.Float64Var(&command.distance, "d", command.distance, "distance between milestone (in meter)")
	flag.BoolVar(&command.reverse, "r", command.reverse, "create milestone in reverse way")
	flag.StringVar(&command.template, "n", command.template, "template of milestone name")
	flag.StringVar(&server.webroot, "w", server.webroot, "web root dir")
	flag.Parse()
	var err error
	if daemon {
		err = server.Run()
	} else {
		err = command.Run()
	}
	if err != nil {
		os.Exit(1)
	}
}

type Command struct {
	template string
	distance float64
	reverse  bool
}

func (c *Command) Run() error {
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
	tmpl, err := template.New("").Parse(c.template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse template: %s\n", err.Error())
		return err
	}
	marker := &milestone.Marker{
		Distance:     c.distance,
		NameTemplate: tmpl,
		Reverse:      c.reverse,
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

type Server struct {
	webroot string
	port    int
}

func (s *Server) Run() error {
	r := mux.NewRouter()
	sub := r.PathPrefix("/cgi").Subrouter()
	sub.HandleFunc("/milestones", milestonesHandler).Methods("POST")
	r.NotFoundHandler = http.FileServer(http.Dir(s.webroot))
	log.Printf("Listening port %d...", s.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen %d: %s\n", s.port, err.Error())
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
