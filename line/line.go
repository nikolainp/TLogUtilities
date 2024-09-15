package main

import (
	"fmt"
	"os"
	"os/signal"
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
