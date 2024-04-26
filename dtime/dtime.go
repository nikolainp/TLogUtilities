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
	conf := getConfig(os.Args)
	run(conf, os.Stdin, os.Stdout)
}

func getConfig(args []string) (conf config) {
	if err := conf.init(args); err != nil {
		switch err.(type) {
		case printVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case printUsage:

		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
		os.Exit(0)
	}
	return conf
}

func run(conf config, sIn io.Reader, sOut io.Writer) {
	var stream streamProcessor

	switch conf.getOperation() {
	case operationFilterByTyme:
		// filter by time: start finish edgeType
		filter := new(lineFilter)
		filter.init(conf.filterBeginTime, conf.filterFinishTime, conf.filterEdge)
		stream.init(filter.LineProcessor)
		stream.run(sIn, sOut)

	case operationTimeGapBack:
		// add TIMEGAP TIMEBACK events
		tg := new(timeGap)
		stream.init(tg.lineProcessor)
		stream.run(sIn, sOut)
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
	poolBuf       sync.Pool
	chBuf         chan streamBuffer
	lineProcessor func([]byte, io.Writer)

	bufSize int
}

func (obj *streamProcessor) init(funcProcessor func([]byte, io.Writer)) {
	obj.bufSize = 1024 * 1024

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return &lines
	}}
	obj.chBuf = make(chan streamBuffer, 1)
	obj.lineProcessor = funcProcessor
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

	reader := bufio.NewReaderSize(sIn, obj.bufSize)

	readBuffer := func(buf []byte) int {
		n, err := reader.Read(buf)
		if n == 0 && err == io.EOF {
			return 0
		}

		if isCancel() {
			return 0
		}

		return n
	}

	for {
		buf := obj.poolBuf.Get().(*[]byte)
		if n := readBuffer(*buf); n == 0 {
			break
		} else {
			obj.chBuf <- streamBuffer{buf, n}
		}
	}

	close(obj.chBuf)
}

func (obj *streamProcessor) doWrite(sOut io.Writer) {

	writer := bufio.NewWriterSize(sOut, obj.bufSize*2)
	lastLine := make([]byte, obj.bufSize*2)
	isExistsLastLine := false

	writeBuffer := func(buf []byte, n int) {
		isLastStringFull := bytes.Equal(buf[n-1:n], []byte("\n"))

		bufSlice := bytes.Split(buf[:n], []byte("\n"))
		for i := range bufSlice {
			if i == 0 && isExistsLastLine {
				lastLine = append(lastLine, bufSlice[i]...)
				if len(bufSlice) > 1 {
					obj.lineProcessor(lastLine, writer)
					isExistsLastLine = false
				}
				continue
			}
			if i == len(bufSlice)-1 {
				if !isLastStringFull {
					lastLine = lastLine[0:len(bufSlice[i])]
					nc := copy(lastLine, bufSlice[i])
					if nc != len(bufSlice[i]) {
						panic(0)
					}
					isExistsLastLine = true
				}
				continue
			}

			obj.lineProcessor(bufSlice[i], writer)
		}
	}

	for {
		if buffer, ok := <-obj.chBuf; ok {
			writeBuffer(*(buffer.buf), buffer.len)

			obj.poolBuf.Put(buffer.buf)
		} else {
			if isExistsLastLine {
				obj.lineProcessor(lastLine, writer)
			}
			break
		}

		if isCancel() {
			break
		}
	}
	writer.Flush()
}

///////////////////////////////////////////////////////////////////////////////
