package main

import (
	"context"
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

type mockFileinfo struct {
	name string
	size int64
}

func (obj mockFileinfo) Name() string       { return obj.name }
func (obj mockFileinfo) Size() int64        { return obj.size }
func (obj mockFileinfo) Mode() fs.FileMode  { return 0 }
func (obj mockFileinfo) ModTime() time.Time { return time.Now() }
func (obj mockFileinfo) IsDir() bool        { return false }
func (obj mockFileinfo) Sys() any           { return nil }

func Test_filePathWalker_runPathWalk(t *testing.T) {

	ctx := context.Background()

	tests := []struct {
		name      string
		input     []fs.FileInfo
		output    []string
		wantSize  int64
		wantCount int
		wantErr   bool
	}{
		{name: "test 1", input: []fs.FileInfo{}, output: []string{}, wantSize: 0, wantCount: 0, wantErr: false},
		{name: "test 2",
			input:    []fs.FileInfo{fs.FileInfo(mockFileinfo{})},
			output:   []string{""},
			wantSize: 0, wantCount: 1, wantErr: false},
		{name: "test 3",
			input:    []fs.FileInfo{fs.FileInfo(mockFileinfo{name: "a1"}), fs.FileInfo(mockFileinfo{name: "b2"})},
			output:   []string{"a1", "b2"},
			wantSize: 0, wantCount: 2, wantErr: false},
		{name: "test 3",
			input:    []fs.FileInfo{fs.FileInfo(mockFileinfo{name: "a1", size: 10}), fs.FileInfo(mockFileinfo{name: "b2", size: 20})},
			output:   []string{"a1", "b2"},
			wantSize: 30, wantCount: 2, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctxWithCancel, ctxCancel := context.WithCancel(ctx)
			defer ctxCancel()

			var gotSize int64
			var gotCount int

			worker := func(path string, walkFunc filepath.WalkFunc) error {
				for _, fsFileInfo := range tt.input {
					if err := walkFunc(fsFileInfo.Name(), fsFileInfo, nil); err != nil {
						return err
					}
				}
				return nil
			}
			monitor := func(size int64, count int) {
				gotSize += size
				gotCount += count
			}

			obj := filePathWalker{
				monitor: monitor,
				input:   make(chan string, len(tt.input)+1),
				output:  make(chan string, len(tt.input)+1),
			}

			if err := obj.runPathWalk(ctxWithCancel, "", worker); (err != nil) != tt.wantErr {
				t.Errorf("filePathWalker.runPathWalk() error = %v, wantErr %v", err, tt.wantErr)
			}
			close(obj.input)

			got := make([]string, 0)
			for out := range obj.input {
				got = append(got, out)
			}
			sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("runOutput() = %v,\n want %v", got, tt.output)
			}

			if tt.wantSize != gotSize {
				t.Errorf("filePathWalker.runPathWalk() size = %v, wantSize %v", gotSize, tt.wantSize)
			}
			if tt.wantCount != gotCount {
				t.Errorf("filePathWalker.runPathWalk() count = %v, wantCount %v", gotCount, tt.wantCount)
			}
		})
	}
}

func Test_filePathWalker_runOutput(t *testing.T) {

	ctx := context.Background()

	tests := []struct {
		name  string
		input []string
	}{
		{"test 1", []string{}},
		{"test 2", []string{"a1"}},
		{"test 3", []string{"a1", "b2"}},
		{"test 4", []string{"a1", "b2", "c3"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctxWithCancel, ctxCancel := context.WithCancel(ctx)
			defer ctxCancel()

			obj := filePathWalker{
				input:  make(chan string, 1),
				output: make(chan string, 1),
			}

			go obj.runOutput(ctxWithCancel)

			for _, in := range tt.input {
				obj.input <- in
			}
			close(obj.input)

			got := make([]string, 0)
			for out := range obj.output {
				got = append(got, out)
			}
			sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

			ctxCancel()

			if !reflect.DeepEqual(got, tt.input) {
				t.Errorf("runOutput() = %v,\n want %v", got, tt.input)
			}
		})
	}
}
