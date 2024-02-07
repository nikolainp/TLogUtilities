package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func Test_streamProcessor(t *testing.T) {
	var bufferIn bytes.Buffer

	for i := 0; i < 1000; i++ {
		bufferIn.WriteString("32:47.733013-0,EXCP,1\n")
	}

	wantSOut := bufferIn.String()
	streamIn := strings.NewReader(wantSOut)
	streamOut := &bytes.Buffer{}

	var obj streamProcessor

	obj.init(func(buf []byte, sOut io.Writer) {
		sOut.Write(buf)
	})
	for i := 0; i < 1; i++ {
		obj.doRead(streamIn)
		obj.doWrite(streamOut)

		if gotSOut := streamOut.String(); gotSOut != wantSOut {
			t.Errorf("streamProcessor() = %v, want %v", gotSOut, wantSOut)
		}

		streamOut.Reset()
	}
}
