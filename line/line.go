package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	version = "dev"
	//	commit  = "none"
	date = "unknown"
)

var cancelChan chan bool

func init() {
	signChan := make(chan os.Signal, 10)
	cancelChan = make(chan bool, 1)

	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		signal := <-signChan
		// Run Cleanup
		fmt.Fprintf(os.Stderr, "\nCaptured %v, stopping and exiting...\n", signal)
		cancelChan <- true
		close(cancelChan)
		os.Exit(0)
	}()
}

func main() {
	var worker pathWalker

	conf := getConfig(os.Args)

	worker.init(conf.isNeedPrefix)
	for _, path := range conf.paths {
		worker.pathWalk(path)

		if isCancel() {
			break
		}
	}
}

func getConfig(args []string) (conf config) {
	if err := conf.init(args); err != nil {
		switch err := err.(type) {
		case printVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case printUsage:
			fmt.Fprint(os.Stderr, err.usage)
		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
		os.Exit(0)
	}
	return conf
}

///////////////////////////////////////////////////////////////////////////////

func isCancel() bool {
	select {
	case _, ok := <-cancelChan:
		return !ok
	default:
		return false
	}
}

///////////////////////////////////////////////////////////////////////////////

type pathWalker struct {
	rootPath string
	check    StreamProcessor

	isNeedPrefix bool
}

func (obj *pathWalker) init(isNeedPrefix bool) {
	obj.rootPath, _ = os.Getwd()
	obj.check = NewStreamProcessor(nil)

	obj.isNeedPrefix = isNeedPrefix
}

func (obj *pathWalker) pathWalk(basePath string) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		obj.doProcess(path)

		if isCancel() {
			return fmt.Errorf("process is cancel")
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *pathWalker) doProcess(fileName string) {
	var subFileName string
	var err error

	if obj.isNeedPrefix {
		subFileName, err = filepath.Rel(obj.rootPath, fileName)
		if err != nil {
			subFileName = fileName
		}
	}

	fileStream, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error open: %q: %v\n", fileName, err)
	}
	defer fileStream.Close()
	obj.check.Run(context.TODO(), subFileName, fileStream, os.Stdout)
}

///////////////////////////////////////////////////////////////////////////////
