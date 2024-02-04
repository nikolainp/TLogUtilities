package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
		//os.Exit(1)
	}()
}

func main() {
	var conf config

	conf.init(os.Args)

	switch conf.getOperation() {
	case operationFilterByTyme:
	// filter by time: start finish edgeType

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
	buf *[]byte
	len int
}

type streamProcessor struct {
	poolBuf sync.Pool
	chBuf   chan streamBuffer

	bufSize int
}

func (obj *streamProcessor) init() {
	obj.bufSize = 1024 * 1024

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return lines
	}}

	obj.chBuf = make(chan streamBuffer, 1)
}

func (obj *streamProcessor) doRead(sIn io.Reader) {
	buf := obj.poolBuf.Get().([]byte)

	reader := bufio.NewReaderSize(sIn, obj.bufSize)
	n, err := reader.Read(buf)
	if n == 0 && err == io.EOF {
		return
	}

	obj.chBuf <- streamBuffer{&buf, n}
}

func (obj *streamProcessor) doWrite(sOut io.Writer) {

	writer := bufio.NewWriterSize(sOut, obj.bufSize*2)

	checkError := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: ", err)
		}
	}
	writeLine := func(line []byte) {
		_, err := writer.Write(line)
		checkError(err)
	}

	buffer := <-obj.chBuf
	bufSlice := (*buffer.buf)[:buffer.len]

	writeLine(bufSlice)
	writer.Flush()

	obj.poolBuf.Put(*buffer.buf)
}

///////////////////////////////////////////////////////////////////////////////
