package main

import (
	"context"
	"fmt"
	"io"
	"os"
)

type FileQueue interface {
	Run(context.Context)
	Pop() (string, io.ReadCloser, bool)
}

func NewFileQueue(files <-chan string) FileQueue {
	return &fileQueue{
		files: files,
		queue: make(chan fileQueueItem, 10),
	}
}

///////////////////////////////////////////////////////////////////////////////

type fileQueue struct {
	files <-chan string
	queue chan fileQueueItem
}

func (obj *fileQueue) Run(ctx context.Context) {
	obj.readFiles(ctx)
}

func (obj *fileQueue) Pop() (string, io.ReadCloser, bool) {
	item, ok := <-obj.queue
	return item.name, item.stream, ok
}

///////////////////////////////////////////////////////////////////////////////

type fileQueueItem struct {
	name   string
	stream io.ReadCloser
}

func (obj *fileQueue) readFiles(ctx context.Context) {
	defer close(obj.queue)

	for isBreak := false; !isBreak; {
		select {
		case fileName, ok := <-obj.files:
			if ok {
				fileStream, err := os.Open(fileName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error open: %q: %v\n", fileName, err)
				}
				obj.queue <- fileQueueItem{name: fileName, stream: fileStream}
			} else {
				isBreak = true
			}
		case <-ctx.Done():
			fmt.Fprint(os.Stderr, "queue stop\n")
			isBreak = true
		}
	}
}
