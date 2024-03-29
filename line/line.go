package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
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
	var printVersion, stripOutput bool
	var worker pathWalker

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.BoolVar(&printVersion, "v", false, "print version")
	fs.BoolVar(&stripOutput, "s", false, "without filename in line")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(fs.Output(), "Usage of %s:\n", os.Args[0])
		fs.PrintDefaults()
		return
	}

	if printVersion {
		fmt.Printf("Version: %s (%s)\n", version, date)
		return
	}

	worker.init(!stripOutput)
	for _, path := range fs.Args() {
		worker.pathWalk(path)

		if isCancel() {
			break
		}
	}
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

type lineChecker struct {
	bufSize int
	buffer  []byte
}

func (obj *lineChecker) init() {
	obj.bufSize = 1024 * 1024 * 10
	obj.buffer = make([]byte, obj.bufSize)
}

func (obj *lineChecker) processStream(sName string, sIn io.Reader, sOut io.Writer) {

	reader := bufio.NewReaderSize(sIn, obj.bufSize)
	writer := bufio.NewWriterSize(sOut, obj.bufSize*2)

	checkError := func(err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: ", err)
		}
	}
	writeLine := func(prefix, line []byte) {
		_, err := writer.Write(prefix)
		checkError(err)
		_, err = writer.Write(line)
		checkError(err)
	}
	getPrefix := func(prefix string) string {
		if len(prefix) == 0 {
			return ""
		}
		return prefix + ":"
	}

	prefixFirstLine := []byte(getPrefix(sName))
	prefixSecondLine := []byte("<line>")

	if isCancel() {
		return
	}

	// read first line
	n, err := reader.Read(obj.buffer)
	if n == 0 && err == io.EOF {
		return
	}
	checkError(err)

	bufSlice := bytes.Split(obj.buffer[:n], []byte("\n"))
	if bytes.Equal(bufSlice[0][:3], []byte("\ufeff")) {
		bufSlice[0] = bufSlice[0][3:]
	}
	if len(bufSlice[0]) == 0 {
		return
	}
	writeLine(prefixFirstLine, bufSlice[0])
	bufSlice = bufSlice[1:]
	isLastLineComplete := bytes.Equal(obj.buffer[n-1:n], []byte("\n"))

	for {
		for i := range bufSlice {
			if obj.isFirstLine(bufSlice[i]) {
				writeLine([]byte{}, []byte("\n"))
				writeLine(prefixFirstLine, bufSlice[i])
			} else {
				writeLine(prefixSecondLine, bufSlice[i])
			}
		}

		err = writer.Flush()
		checkError(err)

		if isCancel() {
			break
		}

		n, err := reader.Read(obj.buffer)
		if n == 0 && err == io.EOF {
			break
		}
		checkError(err)
		bufSlice = bytes.Split(obj.buffer[:n], []byte("\n"))

		if !isLastLineComplete {
			writeLine([]byte{}, bufSlice[0])
			bufSlice = bufSlice[1:]
		}
		isLastLineComplete = bytes.Equal(obj.buffer[n-1:n], []byte("\n"))
	}
	writeLine([]byte{}, []byte("\n"))
	err = writer.Flush()
	checkError(err)
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
