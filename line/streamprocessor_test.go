package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func Test_streamProcessor_Run(t *testing.T) {

	ctx := context.Background()

	tests := []struct {
		name     string
		sIn      string
		wantSOut string
	}{
		{
			"test 1",
			"",
			"\n",
		},
		{
			"test 2",
			`32:47.733006-0,EXCPCNTX,
32:47.733007-0,EXCP,0,
32:47.733013-0,EXCP,1,
32:54.905000-0,EXCP,1,`,
			`test:32:47.733006-0,EXCPCNTX,
test:32:47.733007-0,EXCP,0,
test:32:47.733013-0,EXCP,1,
test:32:54.905000-0,EXCP,1,
`,
		},
		{
			"test 3",
			`32:47.733006-0,EXCPCNTX,0,ClientComputerName=,ServerComputerName=,UserName=,ConnectString=
32:47.733007-0,EXCP,0,process=ragent,OSThread=3668,Exception=81029657-3fe6-4cd6-80c0-36de78fe6657,Descr='src\rtrsrvc\src\remoteinterfaceimpl.cpp(1232):
81029657-3fe6-4cd6-80c0-36de78fe6657:  server_addr=tcp://App:1560 descr=10054(0x00002746): Удаленный хост принудительно разорвал существующее подключение.  line=1582 file=d:\jenkins\ci_builder2\windowsbuild2\platform\src\rtrsrvc\src\dataexchangetcpclientimpl.cpp'
32:47.733013-0,EXCP,1,process=ragent,OSThread=3668,ClientID=6,Exception=NetDataExchangeException,Descr=' server_addr=tcp://App:1541 descr=10054(0x00002746): Удаленный хост принудительно разорвал существующее подключение.  line=1452 file=d:\jenkins\ci_builder2\windowsbuild2\platform\src\rtrsrvc\src\dataexchangetcpclientimpl.cpp
32:54.905000-0,EXCP,1,process=ragent,OSThread=3668,ClientID=4223,Exception=NetDataExchangeException,Descr='server_addr=tcp://App:1541 descr=[fe80::b087:822c:47ce:a93f%13]:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;
[fe80::d1bb:33be:7990:1de2%12]:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;
192.168.7.47:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;
10.10.1.40:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;
 line=1056 file=d:\jenkins\ci_builder2\windowsbuild2\platform\src\rtrsrvc\src\dataexchangetcpclientimpl.cpp'`,
			`test:32:47.733006-0,EXCPCNTX,0,ClientComputerName=,ServerComputerName=,UserName=,ConnectString=
test:32:47.733007-0,EXCP,0,process=ragent,OSThread=3668,Exception=81029657-3fe6-4cd6-80c0-36de78fe6657,Descr='src\rtrsrvc\src\remoteinterfaceimpl.cpp(1232):<line>81029657-3fe6-4cd6-80c0-36de78fe6657:  server_addr=tcp://App:1560 descr=10054(0x00002746): Удаленный хост принудительно разорвал существующее подключение.  line=1582 file=d:\jenkins\ci_builder2\windowsbuild2\platform\src\rtrsrvc\src\dataexchangetcpclientimpl.cpp'
test:32:47.733013-0,EXCP,1,process=ragent,OSThread=3668,ClientID=6,Exception=NetDataExchangeException,Descr=' server_addr=tcp://App:1541 descr=10054(0x00002746): Удаленный хост принудительно разорвал существующее подключение.  line=1452 file=d:\jenkins\ci_builder2\windowsbuild2\platform\src\rtrsrvc\src\dataexchangetcpclientimpl.cpp
test:32:54.905000-0,EXCP,1,process=ragent,OSThread=3668,ClientID=4223,Exception=NetDataExchangeException,Descr='server_addr=tcp://App:1541 descr=[fe80::b087:822c:47ce:a93f%13]:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;<line>[fe80::d1bb:33be:7990:1de2%12]:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;<line>192.168.7.47:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;<line>10.10.1.40:1541:10061(0x0000274D): Подключение не установлено, т.к. конечный компьютер отверг запрос на подключение. ;<line> line=1056 file=d:\jenkins\ci_builder2\windowsbuild2\platform\src\rtrsrvc\src\dataexchangetcpclientimpl.cpp'
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotSize int64
			var gotCount int

			monitor := func(size int64, count int) {
				gotSize += size
				gotCount += count
			}

			obj := NewStreamProcessor(monitor)

			sOut := &bytes.Buffer{}
			obj.Run(ctx, "test", strings.NewReader(tt.sIn), sOut)
			if gotSOut := sOut.String(); gotSOut != tt.wantSOut {
				t.Errorf("processFile() = %v, want %v", gotSOut, tt.wantSOut)
			}
		})
	}
}

func Test_streamProcessor_SetStreamType(t *testing.T) {
	var obj streamProcessor

	tests := []struct {
		name string
		args streamLineType
		data []byte
		want bool
	}{
		{"test 1", streamTLType, []byte("32:47.733006-0,EXCPCNTX,"), true},
		{"test 2", streamTLType, []byte("2024/11/29-15:41:25.263 [launcher-start-thread (start)](12) I! com.e1c.chassis.config.internal.yaml.YamlPersister"), false},
		{"test 3", streamAnsType, []byte("32:47.733006-0,EXCPCNTX,"), false},
		{"test 4", streamAnsType, []byte("2024/11/29-15:41:25.263 [launcher-start-thread (start)](12) I! com.e1c.chassis.config.internal.yaml.YamlPersister"), true},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			obj.SetStreamType(tt.args)
			if got := obj.isFirstLine(tt.data); got != tt.want {
				t.Errorf("lineChecker.isFirstLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_streamProcessor_isFirstLineTypeTL(t *testing.T) {
	var obj streamProcessor

	tests := []struct {
		name string
		in0  []byte
		want bool
	}{
		{"test 1", []byte(""), false},
		{"test 2", []byte("32:47.733006-0,EXCPCNTX,"), true},
		{"test 3", []byte("81029657-3fe6-4cd6-80c0-36de78fe6657"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := obj.isFirstLineTypeTL(tt.in0); got != tt.want {
				t.Errorf("lineChecker.isFirstLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_streamProcessor_isFirstLineTypeAns(t *testing.T) {
	var obj streamProcessor

	tests := []struct {
		name string
		in0  []byte
		want bool
	}{
		{"test 1", []byte(""), false},
		{"test 2", []byte("32:47.733006-0,EXCPCNTX,"), false},
		{"test 3", []byte("2024/11/29-15:41:25.263 [launcher-start-thread (start)](12) I! com.e1c.chassis.config.internal.yaml.YamlPersister"), true},
		{"test 4", []byte("81029657-3fe6-4cd6-80c0-36de78fe6657"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := obj.isFirstLineTypeAns(tt.in0); got != tt.want {
				t.Errorf("streamProcessor.isFirstLineTypeB() = %v, want %v", got, tt.want)
			}
		})
	}
}

// /////////////////////////////////////////////////////////////////////////////
func Benchmark_processStream(b *testing.B) {

	var bufferIn bytes.Buffer

	for i := 0; i < 1000; i++ {
		bufferIn.WriteString("32:47.733013-0,EXCP,1\n")
	}

	ctx := context.Background()
	streamIn := strings.NewReader(bufferIn.String())
	streamOut := &bytes.Buffer{}
	check := NewStreamProcessor(func(int64, int) {})

	b.SetBytes(streamIn.Size())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		check.Run(ctx, "", streamIn, streamOut)
		streamOut.Reset()
	}
}

func Benchmark_isFirstLine(b *testing.B) {
	var check streamProcessor
	data := []byte(`32:47.733007-0,EXCP,0,process=ragent,OSThread=3668,Exception=81029657-3fe6-4cd6-80c0-36de78fe6657,Descr='src\rtrsrvc\src\remoteinterfaceimpl.cpp(1232):`)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		check.isFirstLine(data)
	}
}
