package main

import (
	"bytes"
	"flag"
	"os"
)

type printUsage struct {
	error
	usage string
}
type printVersion struct {
	error
}

type streamLineType int

const (
	streamNoneType streamLineType = iota
	streamTLType
	streamAnsType
)

type config struct {
	programName string
	rootPath    string

	isShowProgress bool
	isNeedPrefix   bool
	paths          []string
	streamType     streamLineType
}

func (obj *config) init(args []string) (err error) {
	var isPrintVersion, isStripOutput bool
	var isTLType, isAnsType bool

	obj.programName = args[0]
	obj.rootPath, _ = os.Getwd()
	obj.streamType = streamNoneType

	fsOut := &bytes.Buffer{}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(fsOut)
	fs.BoolVar(&isPrintVersion, "v", false, "print version")
	fs.BoolVar(&obj.isShowProgress, "p", false, "shows the progress of data through")
	fs.BoolVar(&isStripOutput, "s", false, "without filename in line")
	fs.BoolVar(&isTLType, "tl", false, "1C:Enterprise format")
	fs.BoolVar(&isAnsType, "ans", false, "1C:Analytic format")

	if err := fs.Parse(args[1:]); err != nil {
		return printUsage{usage: fsOut.String()}
	}

	if isPrintVersion {
		return printVersion{}
	}

	obj.isShowProgress = obj.isShowProgress && ouputIsPiped()
	obj.isNeedPrefix = !isStripOutput
	obj.paths = fs.Args()

	if len(obj.paths) == 0 {
		fs.Usage()
		return printUsage{usage: fsOut.String()}
	}

	if isTLType && isAnsType {
		fs.Usage()
		return printUsage{usage: fsOut.String()}
	}
	if isTLType {
		obj.streamType = streamTLType
	}
	if isAnsType {
		obj.streamType = streamAnsType
	}

	return nil
}

func ouputIsPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fi.Mode() & os.ModeNamedPipe) != 0
}
