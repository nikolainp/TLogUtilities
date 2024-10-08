package main

import (
	"context"
	"fmt"
	"io"
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

// var cancelChan chan bool

func main() {

	ctx, cancel := withSignalNotify(context.Background())
	defer cancel()

	conf := getConfig(os.Args)
	monitor := NewMonitor("Load data: files: %d/%d size: %s/%s time: %s [speed %s/s/%s/s ]")
	walker := NewFilePathWalker(monitor.StartProcessing)
	processor := NewStreamProcessor(monitor.FinishProcessing)
	walker.Add(conf.paths...)

	monitor.Run(ctx)
	queue := NewFileQueue(walker.Run(ctx))
	queue.Run(ctx)

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
			os.Exit(0)
		case <-ctxCancel.Done():
			return
		}
	}()

	return ctxCancel, cancel
}

///////////////////////////////////////////////////////////////////////////////

// type pathWalker struct {
// 	rootPath string
// 	check    StreamProcessor

// 	isNeedPrefix bool
// }

// func (obj *pathWalker) init(isNeedPrefix bool) {
// 	obj.rootPath, _ = os.Getwd()
// 	obj.check = NewStreamProcessor(nil)

// 	obj.isNeedPrefix = isNeedPrefix
// }

// func (obj *pathWalker) pathWalk(basePath string) {
// 	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
// 		if err != nil {
// 			fmt.Fprintf(os.Stderr, "Prevent panic by handling failure accessing a path %q: %v\n", path, err)
// 			return err
// 		}
// 		if info.IsDir() {
// 			return nil
// 		}
// 		obj.doProcess(path)

// 		if isCancel() {
// 			return fmt.Errorf("process is cancel")
// 		}

// 		return nil
// 	})
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", basePath, err)
// 	}
// }

///////////////////////////////////////////////////////////////////////////////
