package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
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

type pathWalker struct {
	rootPath string
	check    lineChecker

	isNeedPrefix bool
}

func (obj *pathWalker) init(isNeedPrefix bool) {
	obj.rootPath, _ = os.Getwd()
	obj.check.init()

	obj.isNeedPrefix = isNeedPrefix
}

func (obj *pathWalker) pathWalk(basePath string) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		obj.doProcess(path)

		if isCancel() {
			return fmt.Errorf("process is cancel")
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *pathWalker) doProcess(fileName string) {
	var subFileName string
	var err error

	if obj.isNeedPrefix {
		subFileName, err = filepath.Rel(obj.rootPath, fileName)
		if err != nil {
			subFileName = fileName
		}
	}

	fileStream, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error open: %q: %v\n", fileName, err)
	}
	defer fileStream.Close()
	obj.check.processStream(subFileName, fileStream, os.Stdout)
}

///////////////////////////////////////////////////////////////////////////////

type streamBuffer struct {
	buf *[]byte
	len int
}

type streamLineType int

const (
	streamNoneType streamLineType = iota
	streamTLType
	streamAnsType
)

type lineChecker struct {
	poolBuf sync.Pool
	chBuf   chan streamBuffer

	bufSize          int
	prefixFirstLine  []byte
	prefixSecondLine []byte
}

func (obj *lineChecker) init() {
	obj.bufSize = 1024 * 1024 * 10

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return &lines
	}}
}

func (obj *lineChecker) processStream(sName string, sIn io.Reader, sOut io.Writer) {
	var wg sync.WaitGroup

	getPrefix := func(prefix string) string {
		if len(prefix) == 0 {
			return ""
		}
		return prefix + ":"
	}
	goFunc := func(work func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			work()
		}()
	}

	obj.prefixFirstLine = []byte(getPrefix(sName))
	obj.prefixSecondLine = []byte("<line>")

	obj.chBuf = make(chan streamBuffer, 1)

	goFunc(func() { obj.doRead(sIn) })
	goFunc(func() { obj.doWrite(sOut) })

	wg.Wait()
}

func (obj *lineChecker) doRead(sIn io.Reader) {

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

func (obj *lineChecker) doWrite(sOut io.Writer) {

	writer := bufio.NewWriterSize(sOut, obj.bufSize*2)
	lastLine := make([]byte, obj.bufSize*2)
	isExistsLastLine := false
	streamType := streamNoneType

	writeBuffer := func(buf []byte, n int) {
		isLastStringFull := bytes.Equal(buf[n-1:n], []byte("\n"))

		bufSlice := bytes.Split(buf[:n], []byte("\n"))

		if streamType == streamNoneType {
			if bytes.Equal(bufSlice[0][:3], []byte("\ufeff")) {
				streamType = streamTLType
				bufSlice[0] = bufSlice[0][3:]
			}

			writeLine(writer, obj.prefixFirstLine, bufSlice[0])
			bufSlice = bufSlice[1:]
		}

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

			if isCancel() {
				return
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

	writeLine(writer, []byte{}, []byte("\n"))
	checkError(writer.Flush())
}

func (obj *lineChecker) lineProcessor(data []byte, writer io.Writer) {

	if obj.isFirstLine(data) {
		writeLine(writer, []byte{}, []byte("\n"))
		writeLine(writer, obj.prefixFirstLine, data)
	} else {
		writeLine(writer, obj.prefixSecondLine, data)
	}
}

func (obj *lineChecker) isFirstLine(data []byte) bool {

	isNumber := func(data byte) bool {
		if data == '0' || data == '1' || data == '2' || data == '3' ||
			data == '4' || data == '5' || data == '6' || data == '7' ||
			data == '8' || data == '9' {
			return true
		}

		return false
	}

	// `^\d\d\:\d\d\.\d{6}\-\d+\,\w+\,`
	if len(data) < 14 {
		return false
	}
	if !isNumber(data[0]) || !isNumber(data[1]) || !isNumber(data[3]) || !isNumber(data[4]) {
		return false
	}
	if data[2] != ':' {
		return false
	}
	if data[5] != '.' {
		return false
	}
	if !isNumber(data[6]) || !isNumber(data[7]) || !isNumber(data[8]) ||
		!isNumber(data[9]) || !isNumber(data[10]) || !isNumber(data[11]) {
		return false
	}
	if data[12] != '-' {
		return false
	}
	if !isNumber(data[13]) {
		return false
	}

	return true
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}
}
func writeLine(writer io.Writer, prefix, line []byte) {
	_, err := writer.Write(prefix)
	checkError(err)
	_, err = writer.Write(line)
	checkError(err)
}
