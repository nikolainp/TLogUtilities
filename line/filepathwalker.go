package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type FilePathWalker interface {
	Add(...string) <-chan string
	Run(context.Context)
}

func NewFilePathWalker(callBack func(int64, int)) FilePathWalker {
	return &filePathWalker{
		monitor: callBack,
		input:   make(chan string),
		output:  make(chan string),
	}
}

///////////////////////////////////////////////////////////////////////////////

type filePathWalker struct {
	monitor   func(int64, int)
	rootPaths []string

	bufPaths []string

	input  chan string
	output chan string
}

func (obj *filePathWalker) Add(path ...string) <-chan string {
	obj.rootPaths = append(obj.rootPaths, path...)

	return obj.output
}

func (obj *filePathWalker) Run(ctx context.Context) {

	var wg sync.WaitGroup
	defer wg.Wait()

	goFunc(&wg, func() { obj.runWalk(ctx) })
	obj.runOutput(ctx)

}

///////////////////////////////////////////////////////////////////////////////

func (obj *filePathWalker) runWalk(ctx context.Context) {
	defer close(obj.input)

	for _, path := range obj.rootPaths {
		if err := obj.runPathWalk(ctx, path, filepath.Walk); err != nil {
			//fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", path, err)
			break
		}
	}
}

func (obj *filePathWalker) runOutput(ctx context.Context) {
	defer close(obj.output)

	for isBreak := false; !isBreak; {
		if len(obj.bufPaths) == 0 {
			if path, ok := <-obj.input; ok {
				obj.bufPaths = append(obj.bufPaths, path)
			} else {
				isBreak = true
			}
			continue
		}

		select {
		case path, ok := <-obj.input:
			if ok {
				obj.bufPaths = append(obj.bufPaths, path)
			} 
		case obj.output <- obj.bufPaths[0]:
			obj.bufPaths = obj.bufPaths[1:]
		case <-ctx.Done():
			isBreak = true
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

func (obj *filePathWalker) runPathWalk(ctx context.Context, path string, worker func(string, filepath.WalkFunc) error) error {

	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		select {
		case obj.input <- path:
			obj.monitor(info.Size(), 1)
		case <-ctx.Done():
			return fmt.Errorf("process is cancel")
		}

		return nil
	}

	return worker(path, walkFunc)
}
