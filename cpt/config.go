package main

import (
	"flag"
	"fmt"
	"os"
)

type printUsage struct {
	error
}
type printVersion struct {
	error
}

type config struct {
	programName string

	sourceFolder      string
	destinationFolder string
	fileNames         []string

	transferType dataTransferType
}

func (obj *config) init(args []string) (err error) {
	var flagPrintVersion bool
	var flagMove bool

	obj.programName = args[0]

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&flagPrintVersion, "v", false, "print version")
	fs.BoolVar(&flagMove, "m", false, "move files")

	if err := fs.Parse(args[1:]); fs.NArg() < 2 || err != nil {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
		return printUsage{error: err}
	}

	if flagPrintVersion {
		return printVersion{}
	}

	if flagMove {
		obj.transferType = dataMove
	} else {
		obj.transferType = dataCopy
	}

	obj.sourceFolder = fs.Arg(0)
	obj.destinationFolder = fs.Arg(fs.NArg() - 1)
	for i := 1; i < fs.NArg()-1; i++ {
		obj.fileNames = append(obj.fileNames, fs.Arg(i))
	}

	return nil
}
