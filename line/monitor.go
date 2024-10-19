package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Monitor interface {
	StartProcessing(int64, int)
	FinishProcessing(int64, int)
	Run(context.Context)
}

func NewMonitor(fmt string) Monitor {
	return &monitor{fmtShowState: fmt + "\r"}
}

///////////////////////////////////////////////////////////////////////////////

type monitor struct {
	fmtShowState string
	showStateLen int

	runTime                  time.Time
	totalSize, currentSize   int64
	totalCount, currentCount int

	// mu sync.Mutex
}

func (obj *monitor) StartProcessing(size int64, count int) {
	// obj.mu.Lock()
	// defer obj.mu.Unlock()

	obj.totalSize += size
	obj.totalCount += count
}

func (obj *monitor) FinishProcessing(size int64, count int) {
	// obj.mu.Lock()
	// defer obj.mu.Unlock()

	obj.currentSize += size
	obj.currentCount += count
}

func (obj *monitor) Run(ctx context.Context) {
	obj.runTime = time.Now()
	obj.totalSize = 0
	obj.currentSize = 0
	obj.totalCount = 0
	obj.currentCount = 0

	obj.showState(ctx)
}

///////////////////////////////////////////////////////////////////////////////

func (obj *monitor) showState(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var curDuration time.Duration
	var curSize int64

	for isBreak := false; !isBreak; {
		select {
		case <-ticker.C:
			state := obj.getState()
			obj.showStateLen = state.show(os.Stderr, obj.fmtShowState, &curDuration, &curSize)

		case <-ctx.Done():
			//fmt.Fprint(os.Stderr, "monitor stop\n")
			isBreak = true
		}
	}

	state := obj.getState()
	state.show(os.Stderr, obj.fmtShowState, &curDuration, &curSize)
	fmt.Fprint(os.Stderr, "\n")
}

func (obj *monitor) getState() monitorState {
	res := monitorState{
		showStateLen: obj.showStateLen,
		duration:     time.Since(obj.runTime)}

	// obj.mu.Lock()
	// defer obj.mu.Unlock()

	res.totalSize = obj.totalSize
	res.totalCount = obj.totalCount

	res.currentSize = obj.currentSize
	res.currentCount = obj.currentCount

	return res
}

///////////////////////////////////////////////////////////////////////////////

type monitorState struct {
	duration                 time.Duration
	totalSize, currentSize   int64
	totalCount, currentCount int

	showStateLen int
}

func (obj *monitorState) show(out io.Writer, fmtState string, curDuration *time.Duration, curSize *int64) int {
	var totalSpeed, currentSpped int64

	if obj.duration.Milliseconds() > 0 {
		totalSpeed = 1000 * obj.currentSize / obj.duration.Milliseconds()
	}

	deltaDuration := obj.duration - *curDuration
	if deltaDuration.Milliseconds() > 0 {
		currentSpped = 1000 * (obj.currentSize - *curSize) / deltaDuration.Milliseconds()
		if deltaDuration.Seconds() < 1 {
			currentSpped = 1000 * currentSpped / deltaDuration.Milliseconds()
		}
	}
	if deltaDuration.Seconds() > 1 {
		*curDuration = obj.duration
		*curSize = obj.currentSize
	}

	fmt.Fprint(out, strings.Repeat(" ", obj.showStateLen), "\r")
	n, _ := fmt.Fprintf(out,
		fmtState,
		obj.currentCount, obj.totalCount,
		byteCount(obj.currentSize), byteCount(obj.totalSize),
		obj.duration.Truncate(time.Second),
		byteCount(currentSpped), byteCount(totalSpeed))

	return max(obj.showStateLen, n)
}

///////////////////////////////////////////////////////////////////////////////

func byteCount(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%db", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cb",
		float64(b)/float64(div), "kMGTPE"[exp])
}
