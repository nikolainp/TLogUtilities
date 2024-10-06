package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type FilePathWalker interface {
	Add(...string)
	Run(context.Context) <-chan string
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

func (obj *filePathWalker) Add(path ...string) {
	obj.rootPaths = append(obj.rootPaths, path...)
}

func (obj *filePathWalker) Run(ctx context.Context) <-chan string {

	go obj.runWalk(ctx)
	go obj.runOutput(ctx)

	return obj.output
}

///////////////////////////////////////////////////////////////////////////////

func (obj *filePathWalker) runWalk(ctx context.Context) {
	defer close(obj.input)

	for _, path := range obj.rootPaths {
		if err := obj.runPathWalk(ctx, path, filepath.Walk); err != nil {
			fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", path, err)
		}
	}
}

func (obj *filePathWalker) runOutput(ctx context.Context) {
	defer close(obj.output)

	isDone := false
	for isBreak := false; !isBreak; {
		select {
		case path, ok := <-obj.input:
			if ok {
				obj.bufPaths = append(obj.bufPaths, path)
			} else {
				isDone = true
			}
		case <-ctx.Done():
			isBreak = true
		}

		if len(obj.bufPaths) == 0 {
			isBreak = isDone
		} else {
			obj.output <- obj.bufPaths[0]
			obj.bufPaths = obj.bufPaths[1:]
		}
	}
}

///////////////////////////////////////////////////////////////////////////////

func (obj *filePathWalker) runPathWalk(ctx context.Context, path string, worker func(string, filepath.WalkFunc) error) error {

	walkFunc := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			//fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", path, err)
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
