package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	version = "dev"
	//commit  = "none"
	date = "unknown"
)

type dataTransferType int

const (
	dataCopy dataTransferType = iota
	dataMove
)

func main() {
	conf := getConfig(os.Args)

	var pw fileProcessor
	pw.init()
	pw.pathWalk(conf.sourceFolder)
}

func getConfig(args []string) (conf config) {
	if err := conf.init(args); err != nil {
		switch err.(type) {
		case printVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case printUsage:

		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
		os.Exit(0)
	}
	return conf
}

///////////////////////////////////////////////////////////////////////////////

type fileProcessor struct {
}

func (obj *fileProcessor) init() {
}

func (obj *fileProcessor) pathWalk(basePath string) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		obj.doProcess(path)

		// if isCancel() {
		// 	return fmt.Errorf("process is cancel")
		// }

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *fileProcessor) doProcess(fileName string) {
}

///////////////////////////////////////////////////////////////////////////////
