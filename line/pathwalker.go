package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type Monitor interface {
	Start()
	WriteEvent(frmt string, args ...any)
	NewData(count int, size int64)
	ProcessedData(count int, size int64)
	Stop()
}

type pathWalker struct {
	rootPath string
	monitor  Monitor
	check    lineChecker

	isNeedPrefix bool
}

func (obj *pathWalker) init(isNeedPrefix bool) {
	obj.rootPath, _ = os.Getwd()

	obj.monitor = NewMonitor(cancelChan)
	obj.check.init(obj.monitor)

	obj.isNeedPrefix = isNeedPrefix
}

func (obj *pathWalker) pathWalk(basePath string) {
	defer obj.monitor.Stop()
	obj.monitor.Start()

	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			obj.monitor.WriteEvent("Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}

		obj.monitor.NewData(1, info.Size())
		obj.doProcess(path)

		if isCancel() {
			return fmt.Errorf("process is cancel")
		}

		return nil
	})
	if err != nil {
		obj.monitor.WriteEvent("Error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *pathWalker) doProcess(fileName string) {
	var subFileName string
	var err error

	if obj.isNeedPrefix {
		subFileName, err = filepath.Rel(obj.rootPath, fileName)
		if err != nil {
			subFileName = fileName
		}
	}

	fileStream, err := os.Open(fileName)
	if err != nil {
		obj.monitor.WriteEvent("Error open: %q: %v\n", fileName, err)
	}
	defer fileStream.Close()
	obj.check.processStream(subFileName, fileStream, os.Stdout)
}
