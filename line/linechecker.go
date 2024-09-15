package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

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

	monitor Monitor

	bufSize          int
	prefixFirstLine  []byte
	prefixSecondLine []byte
}

func (obj *lineChecker) init(monitor Monitor) {
	obj.bufSize = 1024 * 1024 * 10

	obj.monitor = monitor

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
			obj.monitor.ProcessedData(1, 0)
			break
		} else {
			obj.monitor.ProcessedData(0, int64(n))
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
			if 3 <= len(bufSlice[0]) &&
				bytes.Equal(bufSlice[0][:3], []byte("\ufeff")) {
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
