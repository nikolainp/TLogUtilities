package main

import (
	"context"
	"sync"
	"time"
)

type Monitor interface {
	StartProcessing(int64, int)
	FinishProcessing(int64, int)
	Run(context.Context)
}

func NewMonitor() Monitor {
	return &monitor{}
}

///////////////////////////////////////////////////////////////////////////////

type monitor struct {
	runTime                  time.Time
	totalSize, currentSize   int64
	totalCount, currentCount int

	mtx sync.Mutex
}

func (obj *monitor) StartProcessing(size int64, count int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.totalSize += size
	obj.totalCount += count
}

func (obj *monitor) FinishProcessing(size int64, count int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	obj.currentSize += size
	obj.currentCount += count
}

func (obj *monitor) Run(context.Context) {
	obj.runTime = time.Now()
}

///////////////////////////////////////////////////////////////////////////////



func (obj *monitor) getState() (totalSize int64, totalCount int, currentSize int64, currentCount int) {
	obj.mtx.Lock()
	defer obj.mtx.Unlock()

	return obj.totalSize, obj.totalCount, obj.currentSize, obj.currentCount
}


