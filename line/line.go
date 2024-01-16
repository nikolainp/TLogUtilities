package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var worker pathWalker

	worker.init()
	for _, path := range os.Args[1:] {
		worker.pathWalk(path)
	}
}

///////////////////////////////////////////////////////////////////////////////

type pathWalker struct {
	check lineChecker
}

func (obj *pathWalker) init() {
	obj.check.init()
}

func (obj *pathWalker) pathWalk(basePath string) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		obj.doProcess(basePath, path)

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *pathWalker) doProcess(basePath string, fileName string) {
	var subFileName string
	var err error

	if strings.Compare(basePath, fileName) == 0 {
		rootPath, _ := os.Getwd()
		subFileName, err = filepath.Rel(rootPath, fileName)
	} else {
		subFileName, err = filepath.Rel(basePath, fileName)
	}
	if err != nil {
		subFileName = fileName
	}

	fileStream, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", fileName, err)
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
	writer := bufio.NewWriterSize(sOut, obj.bufSize)

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

	prefixFirstLine := []byte(sName + ":")
	prefixSecondLine := []byte("<line>")

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
