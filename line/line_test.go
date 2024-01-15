package main

import (
	"bytes"
	"strings"
	"testing"
)

func Test_processStream(t *testing.T) {
	var obj lineChecker
	obj.init()

	tests := []struct {
		name     string
		sIn      string
		wantSOut string
	}{
		{
			"test 1",
			"",
			"",
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
			sOut := &bytes.Buffer{}
			obj.processStream("test", strings.NewReader(tt.sIn), sOut)
			if gotSOut := sOut.String(); gotSOut != tt.wantSOut {
				t.Errorf("processFile() = %v, want %v", gotSOut, tt.wantSOut)
			}
		})
	}
}

func Test_lineChecker_isFirstLine(t *testing.T) {
	var obj lineChecker

	tests := []struct {
		name string
		in0  []byte
		want bool
	}{
		{"test 1", []byte(""), false},
		{"test 2", []byte("32:47.733006-0,EXCPCNTX,"), true},
		{"test 3", []byte("81029657-3fe6-4cd6-80c0-36de78fe6657"), false},
	}

	obj.init()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := obj.isFirstLine(tt.in0); got != tt.want {
				t.Errorf("lineChecker.isFirstLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

// /////////////////////////////////////////////////////////////////////////////
func Benchmark_processStream(b *testing.B) {

	var bufferIn bytes.Buffer
	var check lineChecker

	for i := 0; i < 1000; i++ {
		bufferIn.WriteString("32:47.733013-0,EXCP,1")
	}

	streamIn := strings.NewReader(bufferIn.String())
	streamOut := &bytes.Buffer{}
	check.init()

	b.SetBytes(streamIn.Size())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check.processStream("", streamIn, streamOut)
	}
}

func Benchmark_isFirstLine(b *testing.B) {
	var check lineChecker
	data := []byte(`32:47.733007-0,EXCP,0,process=ragent,OSThread=3668,Exception=81029657-3fe6-4cd6-80c0-36de78fe6657,Descr='src\rtrsrvc\src\remoteinterfaceimpl.cpp(1232):`)

	check.init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		check.isFirstLine(data)
	}
}
