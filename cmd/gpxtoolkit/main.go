package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"gpxtoolkit/cmd"
)

var args = &struct {
	help bool
}{}

var baseCommand = cmd.GPXCommand{
	Output: os.Stdout,
}

var commands = map[string]cmd.Command{
	"time":    &cmd.RewriteTime{GPXCommand: baseCommand},
	"stat":    &cmd.Statistics{GPXCommand: baseCommand},
	"elev":    &cmd.CorrectElevation{GPXCommand: baseCommand},
	"slice":   &cmd.SliceByWaypoints{GPXCommand: baseCommand},
	"project": &cmd.ProjectWaypoints{GPXCommand: baseCommand},
}

func usage(w io.Writer, progname, cmdname string, command cmd.Command, fmtsrt string, args ...interface{}) {
	if command == nil {
		fmt.Fprintf(w, "%s [args...] <command>\n", progname)
		flag.PrintDefaults()
		names := []string{}
		for k := range commands {
			names = append(names, k)
		}
		fmt.Fprintf(w, "\nAvailable commands: %s\n", strings.Join(names, ", "))
	} else {
		command.Usage(w, progname, cmdname)
	}
	if fmtsrt != "" {
		fmt.Fprintf(os.Stderr, "\nERROR: %s\n", fmt.Sprintf(fmtsrt, args...))
	}
}

func main() {
	flag.BoolVar(&args.help, "h", args.help, "help")
	flag.Parse()
	out := flag.CommandLine.Output()
	progname := filepath.Base(os.Args[0])
	cmdname := flag.Arg(0)
	if args.help && cmdname == "" {
		usage(out, progname, "", nil, "")
		return
	}
	if cmdname == "" {
		usage(out, progname, "", nil, "Missing command")
		os.Exit(int(syscall.EINVAL))
	}
	cmd := commands[cmdname]
	if cmd == nil {
		usage(out, progname, "", nil, "Unknown command: %s", cmdname)
		os.Exit(int(syscall.EINVAL))
	}
	if args.help {
		cmd.Usage(out, progname, flag.Arg(0))
		return
	}
	if len(flag.Args())-1 < cmd.MinArguments() {
		usage(out, progname, cmdname, cmd, "Not enough argument(s) for '%s': expect=%d, actual=%d", cmdname, cmd.MinArguments(), len(flag.Args())-1)
		os.Exit(int(syscall.EINVAL))
	}
	err := cmd.Run(progname, cmdname, flag.Args()[1:])
	if err != nil {
		log.Printf("Command '%s' failed: %s", cmdname, err.Error())
	}
}
