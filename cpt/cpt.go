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

	var fp fileProcessor
	fp.init(conf.sourceFolder, conf.destinationFolder)

	pathWalk(conf.sourceFolder, fp)
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

func pathWalk(basePath string, fp fileProcessor) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		fp.doProcess(path)

		// if isCancel() {
		// 	return fmt.Errorf("process is cancel")
		// }

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", basePath, err)
	}
}

///////////////////////////////////////////////////////////////////////////////

type fileProcessor struct {
	source      string
	destination string
}

func (obj *fileProcessor) init(source, destination string) {
	obj.source = source
	obj.destination = destination
}

func (obj *fileProcessor) doProcess(fileName string) {

	getSubFilePath := func(path string) string {
		dir, file := filepath.Split(path)
		base := filepath.Base(dir)
		return filepath.Join(base, file)
	}

	subFilePath := getSubFilePath(fileName)

	fmt.Printf("mv %s %s\n", fileName,
		filepath.Join(obj.destination, subFilePath))

}

///////////////////////////////////////////////////////////////////////////////
