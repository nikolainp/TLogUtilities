package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type monitor struct {
	partsTotal    int
	partsFinished int
	sizeTotal     int64
	sizeFinished  int64
	messageBuffer []string

	fmtShowProgress string
	isShowProgress  bool

	startTime time.Time
	ticker    *time.Ticker
	done      chan bool

	mu         sync.Mutex
	wg         sync.WaitGroup
	cancelChan chan bool
}

func NewMonitor(isCancelChan chan bool) Monitor {
	obj := new(monitor)
	obj.done = make(chan bool)

	obj.fmtShowProgress = "files: %d/%d size: %s/%s time: %s [speed %s/s/%s/s ]                           \r"
	obj.isShowProgress = ouputIsPiped()

	obj.cancelChan = isCancelChan

	obj.Start()

	return obj
}

func (obj *monitor) WriteEvent(frmt string, args ...any) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.messageBuffer = append(obj.messageBuffer, fmt.Sprintf(frmt, args...))
}

func (obj *monitor) NewData(count int, size int64) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.partsTotal += count
	obj.sizeTotal += size
}

func (obj *monitor) ProcessedData(count int, size int64) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.partsFinished += count
	obj.sizeFinished += size
}

func (obj *monitor) Start() {
	obj.Stop()

	obj.partsTotal = 0
	obj.partsFinished = 0
	obj.sizeTotal = 0
	obj.sizeFinished = 0
	//	obj.fmtShowProgress = obj.fmtShowProgress + "          \r"

	obj.print()
}

func (obj *monitor) Stop() {
	if obj.ticker == nil {
		return
	}

	defer obj.ticker.Stop()

	obj.done <- true
	obj.wg.Wait()
}

// func (obj *Monitor) IsCancel() bool {
// 	select {
// 	case _, ok := <-obj.cancelChan:
// 		return !ok
// 	default:
// 		return false
// 	}
// }

// func (obj *Monitor) Cancel() chan bool {
// 	return obj.cancelChan
// }

///////////////////////////////////////////////////////////////////////////////

func (obj *monitor) print() {
	var prevFinishedSize int64
	var prevDuration time.Duration

	obj.startTime = time.Now()
	obj.ticker = time.NewTicker(500 * time.Millisecond)

	doPrintProgress := func() {
		var speed int64
		var totalSpeed int64

		totalDuration := time.Since(obj.startTime)
		if totalDuration.Milliseconds() == 0 {
			return
		}
		totalSpeed = 1000 * obj.sizeFinished / totalDuration.Milliseconds()

		deltaDuration := totalDuration - prevDuration
		if deltaDuration.Milliseconds() > 0 {
			speed = 1000 * (obj.sizeFinished - prevFinishedSize) / deltaDuration.Milliseconds()
			if deltaDuration.Seconds() < 1 {
				speed = 1000 * speed / deltaDuration.Milliseconds()
			}
		}
		if deltaDuration.Seconds() > 1 {
			prevDuration = totalDuration
			prevFinishedSize = obj.sizeFinished
		}

		fmt.Fprintf(os.Stderr,
			obj.fmtShowProgress,
			obj.partsFinished, obj.partsTotal,
			byteCount(obj.sizeFinished), byteCount(obj.sizeTotal),
			totalDuration.Truncate(time.Second),
			byteCount(speed), byteCount(totalSpeed))
	}
	doPrintMessages := func() {
		for i := range obj.messageBuffer {
			fmt.Fprint(os.Stderr, obj.messageBuffer[i])
		}
		obj.messageBuffer = obj.messageBuffer[:0]
	}
	doPrint := func() {
		defer obj.mu.Unlock()
		obj.mu.Lock()

		doPrintMessages()
		if obj.isShowProgress {
			doPrintProgress()
		}
	}

	obj.wg.Add(1)
	go func() {
		defer obj.wg.Done()

		for {
			var done, cancel bool

			select {
			case done = <-obj.done:

			case _, ok := <-obj.cancelChan:
				cancel = !ok

			case <-obj.ticker.C:
				doPrint()
			}

			if done || cancel {
				break
			}
		}

		doPrintProgress()
		fmt.Fprintf(os.Stderr, "\n")
	}()
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

func ouputIsPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fi.Mode() & os.ModeNamedPipe) != 0
}
