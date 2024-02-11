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
	buf *[]byte
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
	buf := obj.poolBuf.Get().([]byte)
	reader := bufio.NewReaderSize(sIn, obj.bufSize)

	for {
		n, err := reader.Read(buf)
		if n == 0 && err == io.EOF {
			break
		}

		obj.chBuf <- streamBuffer{&buf, n}

		if isCancel() {
			break
		}
	}
	close(obj.chBuf)
}

func (obj *streamProcessor) doWrite(sOut io.Writer) {

	writer := bufio.NewWriterSize(sOut, obj.bufSize*2)

	// checkError := func(err error) {
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, "Error: ", err)
	// 	}
	// }
	// writeLine := func(line []byte) {
	// 	_, err := writer.Write(line)
	// 	checkError(err)
	// }

	for buffer := range obj.chBuf {

		for _, buf := range bytes.Split((*buffer.buf)[:buffer.len], []byte("\n")) {
			obj.processor(buf, writer)
		}

		obj.poolBuf.Put(*(buffer.buf))
		writer.Flush()

		if isCancel() {
			break
		}
	}
}

///////////////////////////////////////////////////////////////////////////////
