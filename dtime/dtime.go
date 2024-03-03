package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	version = "dev"
	//commit  = "none"
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
	var conf config

	conf.init(os.Args, version, date)
	run(conf, os.Stdin, os.Stdout)
}

func run(conf config, sIn io.Reader, sOut io.Writer) {
	var stream streamProcessor

	switch conf.getOperation() {
	case operationFilterByTyme:
		// filter by time: start finish edgeType
		filter := new(lineFilter)
		filter.init(conf.filterBeginTime, conf.filterFinishTime, conf.filterEdge)
		stream.init(filter.process)
		stream.run(sIn, sOut)

	case operationTimeGapBack:
		// add TIMEGAP TIMEBACK events
	}

	// operations with time
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

type streamBuffer struct {
	buf []byte
	len int
}

type streamProcessor struct {
	poolBuf   sync.Pool
	chBuf     chan streamBuffer
	processor func([]byte, io.Writer)

	bufSize int
}

func (obj *streamProcessor) init(funcProcessor func([]byte, io.Writer)) {
	obj.bufSize = 1024 * 1024

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return lines
	}}
	obj.chBuf = make(chan streamBuffer, 1)
	obj.processor = funcProcessor
}

func (obj *streamProcessor) run(sIn io.Reader, sOut io.Writer) {
	var wg sync.WaitGroup

	goFunc := func(work func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			work()
		}()
	}

	goFunc(func() { obj.doRead(sIn) })
	goFunc(func() { obj.doWrite(sOut) })

	wg.Wait()
}

func (obj *streamProcessor) doRead(sIn io.Reader) {
	var lastLineIndex, nextLineIndex int

	reader := bufio.NewReaderSize(sIn, obj.bufSize)

	buf := obj.poolBuf.Get().([]byte)
	for {
		n, err := reader.Read(buf[nextLineIndex:])
		n += nextLineIndex
		if n == 0 && err == io.EOF {
			break
		}

		newBuf := obj.poolBuf.Get().([]byte)
		lastLineIndex = bytes.LastIndexByte(buf[:n], '\n')
		if lastLineIndex == -1 {
			nextUntillNewLine, err := reader.ReadBytes('\n') //read entire line
			if err != nil && err != io.EOF {
				break
			}
			buf = append(buf[:n], nextUntillNewLine...)
			nextLineIndex = 0
			n += len(nextUntillNewLine)
		} else {
			copy(newBuf, buf[lastLineIndex+1:n])
			nextLineIndex = n - lastLineIndex - 1
			n = lastLineIndex
		}
		obj.chBuf <- streamBuffer{buf, n}
		buf = newBuf

		if isCancel() {
			break
		}
	}
	close(obj.chBuf)
}

func (obj *streamProcessor) doWrite(sOut io.Writer) {

	writer := bufio.NewWriterSize(sOut, obj.bufSize*2)

	for buffer := range obj.chBuf {

		for _, buf := range bytes.Split(buffer.buf[:buffer.len], []byte("\n")) {
			obj.processor(buf, writer)
		}

		obj.poolBuf.Put(buffer.buf)
		writer.Flush()

		if isCancel() {
			break
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
