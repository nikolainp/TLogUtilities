package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
)

type StreamProcessor interface {
	Run(context.Context, string, io.Reader, io.Writer)
}

func NewStreamProcessor(callBack func(int64, int)) StreamProcessor {

	obj := new(streamProcessor)

	obj.monitor = callBack
	obj.bufSize = 1024 * 1024 * 1

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return &lines
	}}

	return obj
}

///////////////////////////////////////////////////////////////////////////////

type streamProcessor struct {
	monitor func(int64, int)
	poolBuf sync.Pool
	chBuf   chan streamBuffer

	bufSize          int
	prefixFirstLine  []byte
	prefixSecondLine []byte
}

func (obj *streamProcessor) Run(ctx context.Context, sName string, sIn io.Reader, sOut io.Writer) {
	var wg sync.WaitGroup

	getPrefix := func(prefix string) string {
		if len(prefix) == 0 {
			return ""
		}
		return prefix + ":"
	}
	// goFunc := func(work func()) {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		work()
	// 	}()
	// }

	obj.prefixFirstLine = []byte(getPrefix(sName))
	obj.prefixSecondLine = []byte("<line>")

	obj.chBuf = make(chan streamBuffer, 1)

	goFunc(&wg, func() { obj.doRead(ctx, sIn) })
	goFunc(&wg, func() { obj.doWrite(ctx, sOut) })

	wg.Wait()

	obj.monitor(0, 1)
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

func (obj *streamProcessor) doRead(ctx context.Context, sIn io.Reader) {

	reader := bufio.NewReaderSize(sIn, obj.bufSize)

	readBuffer := func(buf []byte) int {
		n, err := reader.Read(buf)
		if n == 0 && err == io.EOF {
			return 0
		}

		return n
	}

	for isBreak := false; !isBreak; {
		buf := obj.poolBuf.Get().(*[]byte)
		if n := readBuffer(*buf); n == 0 {
			isBreak = true
		} else {
			select {
			case obj.chBuf <- streamBuffer{buf, n}:
			case <-ctx.Done():
				fmt.Fprint(os.Stderr, "read stop\n")
				isBreak = true
			}
		}
	}

	close(obj.chBuf)
}

func (obj *streamProcessor) doWrite(ctx context.Context, sOut io.Writer) {

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
			select {
			case <-ctx.Done():
				fmt.Fprint(os.Stderr, "write stop\n")
				return
			default:

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
	}

	for isBreak := false; !isBreak; {
		select {
		case buffer, ok := <-obj.chBuf:
			if ok {
				writeBuffer(*(buffer.buf), buffer.len)

				obj.poolBuf.Put(buffer.buf)
			} else {
				if isExistsLastLine {
					obj.lineProcessor(lastLine, writer)
				}
				isBreak = true
			}
		case <-ctx.Done():
			fmt.Fprint(os.Stderr, "write stop\n")
			isBreak = true
		}
	}

	writeLine(writer, []byte{}, []byte("\n"))
	checkError(writer.Flush())
}

func (obj *streamProcessor) lineProcessor(data []byte, writer io.Writer) {

	obj.monitor(int64(len(data)), 0)

	if obj.isFirstLine(data) {
		writeLine(writer, []byte{}, []byte("\n"))
		writeLine(writer, obj.prefixFirstLine, data)
	} else {
		writeLine(writer, obj.prefixSecondLine, data)
	}
}

func (obj *streamProcessor) isFirstLine(data []byte) bool {

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
