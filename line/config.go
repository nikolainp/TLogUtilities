package main

import (
	"bytes"
	"flag"
)

type printUsage struct {
	error
	usage string
}
type printVersion struct {
	error
}

type config struct {
	programName string

	isNeedPrefix bool
	paths        []string
}

func (obj *config) init(args []string) (err error) {
	var isPrintVersion, stripOutput bool

	obj.programName = args[0]
	fsOut := &bytes.Buffer{}

	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(fsOut)
	fs.BoolVar(&isPrintVersion, "v", false, "print version")
	fs.BoolVar(&stripOutput, "s", false, "without filename in line")

	if err := fs.Parse(args[1:]); err != nil {
		return printUsage{usage: fsOut.String()}
	}

	if isPrintVersion {
		return printVersion{}
	}

	obj.isNeedPrefix = !stripOutput
	obj.paths = fs.Args()

	if len(obj.paths) == 0 {
		fs.Usage()
		return printUsage{usage: fsOut.String()}
	}

	return nil
}
