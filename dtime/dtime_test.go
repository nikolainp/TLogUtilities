package main

import (
	"reflect"
	"testing"
	"time"
)

func Test_lineFilter_init(t *testing.T) {
	type args struct {
		start time.Time
		stop  time.Time
		edge  edgeType
	}
	tests := []struct {
		name string
		args args
		want lineFilter
	}{
		{
			"test1",
			args{time.Date(2024, 2, 1, 12, 30, 0, 1, time.UTC),
				time.Date(2024, 2, 1, 12, 30, 0, 2, time.UTC), edgeStart},
			lineFilter{time.Date(2024, 2, 1, 12, 30, 0, 1, time.UTC),
				time.Date(2024, 2, 1, 12, 30, 0, 2, time.UTC),
				[]byte("24020112.log:30:00.000001"),
				[]byte("24020112.log:30:00.000002"), edgeStart},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got lineFilter
			got.init(tt.args.start, tt.args.stop, tt.args.edge)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFileSearcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lineFilter_isTrueLineByStart(t *testing.T) {
	var filter lineFilter
	filter.init(
		time.Date(2024, 1, 12, 15, 30, 0, 1000, time.Local),
		time.Date(2024, 1, 12, 15, 35, 0, 2000, time.Local),
		edgeStop)

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"test1", []byte(""), false},
		{"test2", []byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,`), true},
		{"test3", []byte(`.\rphost_2345\24011215.log:22:47.733007-0,EXCP,`), false},
		{"test4", []byte(`.\rphost_2345\24011215.log:42:47.733007-0,EXCP,`), false},
		{"test5", []byte(`.\rphost_2345\24011215.log:35:00.000001-0,EXCP,`), true},
		{"test6", []byte(`.\rphost_2345\24011215.log:35:00.000003-0,EXCP,`), false},
		{"test7", []byte(`.\rphost_2345\24011215.log:35:00.000003-1000,EXCP,`), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter.isTrueLineByStart(tt.data); got != tt.want {
				t.Errorf("lineFilter.isTrueLineByStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lineFilter_isTrueLineByStop(t *testing.T) {
	var filter lineFilter
	filter.init(
		time.Date(2024, 1, 12, 15, 30, 0, 1, time.Local),
		time.Date(2024, 1, 12, 15, 35, 0, 2, time.Local),
		edgeStop)

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"test1", []byte(""), false},
		{"test2", []byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,`), true},
		{"test3", []byte(`.\rphost_2345\24011215.log:22:47.733007-0,EXCP,`), false},
		{"test4", []byte(`.\rphost_2345\24011215.log:42:47.733007-0,EXCP,`), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filter.isTrueLineByStop(tt.data); got != tt.want {
				t.Errorf("lineFilter.isTrueLineByStop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getStrTimeFromLine(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want []byte
	}{
		{"test1", []byte(""), nil},
		{
			"test2",
			[]byte(`011215.log:32:47.733007-0,EXCP,0,process=ragent,OSThread=3668,Exception=81029657-3fe6-4cd6-80c0-36de78fe6657,Descr='src\rtrsrvc\src\remoteinterfaceimpl.cpp(1232):`),
			nil,
		},
		{
			"test2",
			[]byte(`.\rphost_2345\24011215.log:32:47.7330`),
			nil,
		},
		{
			"test2",
			[]byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,0,process=ragent,OSThread=3668,Exception=81029657-3fe6-4cd6-80c0-36de78fe6657,Descr='src\rtrsrvc\src\remoteinterfaceimpl.cpp(1232):`),
			[]byte("24011215.log:32:47.733007"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := getStrTimeFromLine(tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getStrTimeFromLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_isTrueLineByStart(b *testing.B) {
	var filter lineFilter
	filter.init(
		time.Date(2024, 1, 12, 15, 30, 0, 1, time.Local),
		time.Date(2024, 1, 12, 15, 35, 0, 2, time.Local),
		edgeStop)
	data := []byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.isTrueLineByStart(data)
	}
}

func Benchmark_isTrueLineByStop(b *testing.B) {
	var filter lineFilter
	filter.init(
		time.Date(2024, 1, 12, 15, 30, 0, 1, time.Local),
		time.Date(2024, 1, 12, 15, 35, 0, 2, time.Local),
		edgeStop)
	data := []byte(`.\rphost_2345\24011215.log:32:47.733007-0,EXCP,`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.isTrueLineByStop(data)
	}
}
