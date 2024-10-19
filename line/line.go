package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

var (
	version = "dev"
	//	commit  = "none"
	date = "unknown"
)

// var cancelChan chan bool

func main() {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := withSignalNotify(context.Background())
	defer cancel()

	conf := getConfig(os.Args)
	monitor := NewMonitor("Load data: files: %d/%d size: %s/%s time: %s [%s/s/%s/s]")
	walker := NewFilePathWalker(monitor.StartProcessing)
	processor := NewStreamProcessor(monitor.FinishProcessing)

	queue := NewFileQueue(walker.Add(conf.paths...))

	if conf.isShowProgress {
		goFunc(&wg, func() { monitor.Run(ctx) })
	}
	goFunc(&wg, func() { walker.Run(ctx) })
	goFunc(&wg, func() { queue.Run(ctx) })

	for isBreak := false; !isBreak; {
		select {
		case <-ctx.Done():
			isBreak = true
			for ok := true; ok; {
				var file io.Closer
				if _, file, ok = queue.Pop(); ok {
					file.Close()
				}
			}
		default:
			name, file, ok := queue.Pop()
			if ok {
				if conf.isNeedPrefix {
					subName, err := filepath.Rel(conf.rootPath, name)
					if err != nil {
						subName = name
					}
					processor.Run(ctx, subName, file, os.Stdout)
				} else {
					processor.Run(ctx, "", file, os.Stdout)
				}
				file.Close()
			} else {
				isBreak = true
			}
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

func withSignalNotify(ctx context.Context) (context.Context, context.CancelFunc) {
	signChan := make(chan os.Signal, 10)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	ctxCancel, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case signal := <-signChan:
			// Run Cleanup
			fmt.Fprintf(os.Stderr, "\nCaptured %v, stopping and exiting...\n", signal)
			cancel()
			//os.Exit(0)
		case <-ctxCancel.Done():
			//fmt.Fprint(os.Stderr, "signal stop\n")
			return
		}
	}()

	return ctxCancel, cancel
}

func goFunc(wg *sync.WaitGroup, do func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		do()
	}()
}

///////////////////////////////////////////////////////////////////////////////
