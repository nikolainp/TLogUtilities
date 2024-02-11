package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func Test_streamProcessor(t *testing.T) {
	var bufferIn bytes.Buffer

	for i := 0; i < 10; i++ {
		bufferIn.WriteString("32:47.733013-0,EXCP,1\n")
	}

	wantSOut := bufferIn.String()
	streamIn := strings.NewReader(wantSOut)
	streamOut := &bytes.Buffer{}

	var obj streamProcessor

	obj.init(func(buf []byte, sOut io.Writer) {
		if len(buf) == 0 {
			return
		}
		sOut.Write(buf)
		sOut.Write([]byte("\n"))
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

func Test_run(t *testing.T) {
	tests := []struct {
		name   string
		conf   config
		strIn  string
		strOut string
	}{
		{"test 1", config{}, "", ""},
		{"test 2",
			config{
				operation:        operationFilterByTyme,
				filterBeginTime:  time.Date(2024, 1, 12, 15, 30, 0, 1000, time.Local),
				filterFinishTime: time.Date(2024, 1, 12, 15, 35, 0, 2000, time.Local),
				filterEdge:       edgeStop,
			},
			`
.\rphost_2345\24011215.log:32:47.733007-0,EXCP
.\rphost_2345\24011215.log:22:47.733007-0,EXCP,
.\rphost_2345\24011215.log:42:47.733007-0,EXCP,
.\rphost_2345\24011215.log:35:00.000001-0,EXCP
.\rphost_2345\24011215.log:35:00.003003-0,EXCP,
.\rphost_2345\24011215.log:35:00.003003-1000,EXCP
`,
			`.\rphost_2345\24011215.log:32:47.733007-0,EXCP
.\rphost_2345\24011215.log:35:00.000001-0,EXCP
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sOut := &bytes.Buffer{}
			run(tt.conf, strings.NewReader(tt.strIn), sOut)
			if gotSOut := sOut.String(); gotSOut != tt.strOut {
				t.Errorf("run() = %v, want %v", gotSOut, tt.strOut)
			}
		})
	}
}

func Test_streamProcessor_run(t *testing.T) {
	type args struct {
		sIn io.Reader
	}
	tests := []struct {
		name     string
		obj      *streamProcessor
		args     args
		wantSOut string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sOut := &bytes.Buffer{}
			tt.obj.run(tt.args.sIn, sOut)
			if gotSOut := sOut.String(); gotSOut != tt.wantSOut {
				t.Errorf("streamProcessor.run() = %v, want %v", gotSOut, tt.wantSOut)
			}
		})
	}
}
