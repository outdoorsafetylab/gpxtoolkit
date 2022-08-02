package cmd

import "io"

type Command interface {
	Usage(w io.Writer, progname, cmdname string)
	MinArguments() int
	Run(progname, cmdname string, args []string) error
}
