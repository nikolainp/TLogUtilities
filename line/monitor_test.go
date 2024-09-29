package main

import (
	"bytes"
	"testing"
	"time"
)

func Test_monitor_StartProcessing(t *testing.T) {
	type args struct {
		size  int64
		count int
	}
	tests := []struct {
		name  string
		obj   *monitor
		input args
		want  args
	}{
		{name: "test 1", obj: &monitor{}, input: args{size: 0, count: 0}, want: args{size: 0, count: 0}},
		{name: "test 2", obj: &monitor{}, input: args{size: 10, count: 0}, want: args{size: 10, count: 0}},
		{name: "test 3", obj: &monitor{}, input: args{size: 0, count: 20}, want: args{size: 0, count: 20}},
		{name: "test 4", obj: &monitor{}, input: args{size: 30, count: 40}, want: args{size: 30, count: 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.obj.StartProcessing(tt.input.size, tt.input.count)

			if tt.obj.totalSize != tt.want.size {
				t.Errorf("monitor.StartProcessing() size = %v, wantSize %v", tt.obj.totalSize, tt.want.size)
			}
			if tt.obj.totalCount != tt.want.count {
				t.Errorf("monitor.StartProcessing() count = %v, wantCount %v", tt.obj.totalCount, tt.want.count)
			}
		})
	}
}

func Test_monitor_FinishProcessing(t *testing.T) {
	type args struct {
		size  int64
		count int
	}
	tests := []struct {
		name  string
		obj   *monitor
		input args
		want  args
	}{
		{name: "test 1", obj: &monitor{}, input: args{size: 0, count: 0}, want: args{size: 0, count: 0}},
		{name: "test 2", obj: &monitor{}, input: args{size: 10, count: 0}, want: args{size: 10, count: 0}},
		{name: "test 3", obj: &monitor{}, input: args{size: 0, count: 20}, want: args{size: 0, count: 20}},
		{name: "test 4", obj: &monitor{}, input: args{size: 30, count: 40}, want: args{size: 30, count: 40}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.obj.FinishProcessing(tt.input.size, tt.input.count)

			if tt.obj.currentSize != tt.want.size {
				t.Errorf("monitor.FinishProcessing() size = %v, wantSize %v", tt.obj.currentSize, tt.want.size)
			}
			if tt.obj.currentCount != tt.want.count {
				t.Errorf("monitor.FinishProcessing() count = %v, wantCount %v", tt.obj.currentCount, tt.want.count)
			}
		})
	}
}

func Test_monitor_showCurrentState(t *testing.T) {
	type args struct {
		curDuration time.Duration
		curSize     int64
	}
	tests := []struct {
		name string
		obj  *monitorState
		args args
		want string
	}{
		{name: "test 1",
			obj: &monitorState{duration: time.Duration(20),
				totalSize: 20, totalCount: 20,
				currentSize: 15, currentCount: 15},
			args: args{curDuration: time.Duration(10), curSize: 10},
			want: "Load data: files: 15/20 size: 15b/20b time: 0s [speed 0b/s/0b/s ]"},
		{name: "test 2",
			obj: &monitorState{duration: time.Duration(20000000),
				totalSize: 20, totalCount: 20,
				currentSize: 15, currentCount: 15},
			args: args{curDuration: time.Duration(10000000), curSize: 10},
			want: "Load data: files: 15/20 size: 15b/20b time: 0s [speed 50.0kb/s/750b/s ]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}

			tt.obj.show(out,
				"Load data: files: %d/%d size: %s/%s time: %s [speed %s/s/%s/s ]",
				&tt.args.curDuration, &tt.args.curSize)

			if gotOut := out.String(); gotOut != tt.want {
				t.Errorf("monitor.showCurrentState() = %v, want %v", gotOut, tt.want)
			}
		})
	}
}
