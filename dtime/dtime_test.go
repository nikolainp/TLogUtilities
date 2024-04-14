package main

import (
	"bytes"
	"os"
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

	obj.init(func(buf []byte) bool {
		return true
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
	tests := []struct {
		name string
		sIn  string
		sOut string
	}{
		{"test 1", "", ""},
		{"test 2",
			`0
01
012
0123
01234
012345
0123456
01234567
012345678
0123456789
0123456789a
0123456789ab
0123456789abc
0123456789abcd
0123456789abcde
0123456789abcdef
0123456789abcdefg`,
			`0
01
012
0123
01234
012345
0123456
01234567
012345678
0123456789
0123456789a
0123456789ab
0123456789abc
0123456789abcd
0123456789abcde
0123456789abcdef
0123456789abcdefg
`,
		},
		{"test 3",
			`0
01
012
0123
01234
012345
0123456
01234567
012345678
0123456789
0123456789a
0123456789ab
0123456789abc
0123456789abcd
0123456789abcde
0123456789abcdef
0123456789abcdefg
`,
			`0
01
012
0123
01234
012345
0123456
01234567
012345678
0123456789
0123456789a
0123456789ab
0123456789abc
0123456789abcd
0123456789abcde
0123456789abcdef
0123456789abcdefg
`,
		},
		{"test 4",
			`0123456789abcdefg
0123456789abcdef
0123456789abcde
0123456789abcd
0123456789abc
0123456789ab
0123456789a
0123456789
012345678
01234567
0123456
012345
01234
0123
012
01
0
`,
			`0123456789abcdefg
0123456789abcdef
0123456789abcde
0123456789abcd
0123456789abc
0123456789ab
0123456789a
0123456789
012345678
01234567
0123456
012345
01234
0123
012
01
0
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var obj streamProcessor

			sOut := &bytes.Buffer{}
			obj.init(func(buf []byte) bool {
				return true
			})
			obj.bufSize = 5

			obj.run(strings.NewReader(tt.sIn), sOut)
			gotSOut := sOut.String()
			if gotSOut != tt.sOut {
				t.Errorf("streamProcessor.run() = %v, want %v", gotSOut, tt.sIn)
			}
		})
	}
}

func Benchmark_Test_doRead(b *testing.B) {
	file, err := os.Open("c:\\temp\\tj.out.txt")
	if err != nil {
		return
	}

	var obj streamProcessor
	obj.init(func(buf []byte) bool {
		return true
	})
	obj.bufSize = 30
	obj.run(file, os.Stdout)
}

func Benchmark_doRead(b *testing.B) {
	strIn := `0
01
012
0123
01234
012345
0123456
01234567
012345678
0123456789
0123456789a
0123456789ab
0123456789abc
0123456789abcd
0123456789abcde
0123456789abcdef
0123456789abcdefg
`

	var obj streamProcessor

	sIn := strings.NewReader(strIn)
	sOut := &bytes.Buffer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.init(func(buf []byte) bool { return true })
		obj.bufSize = 5
		obj.run(sIn, sOut)
		sIn.Reset(strIn)
	}
}
