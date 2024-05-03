package main

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type timeGapThread struct {
	process string
	thread   string
}

type timeGap struct {
	processes map[string]time.Time
	threads    map[timeGapThread]time.Time
}

func (obj *timeGap) init() {
	obj.processes = make(map[string]time.Time)
	obj.threads = make(map[timeGapThread]time.Time)
}

func (obj *timeGap) lineProcessor(data []byte, writer io.Writer) {

	getThread := func(data []byte) []byte {
		threadMask := []byte(",OSThread=")
		threadPosition := bytes.Index(data, threadMask)
		commaPosition := bytes.Index(data[threadPosition+len(threadMask):], []byte(","))

		return data[threadPosition+len(threadMask) : threadPosition+len(threadMask)+commaPosition]
	}
	writeEvent := func(f string, args ...any) []byte {
		return []byte(fmt.Sprintf(f, args...))
	}

	timeBegin, timeFinish := getStrTimePosition(data)
	if timeBegin == 0 && timeFinish == 0 {
		return
	}

	strProcess := data[0:timeBegin]
	strTime := data[timeBegin:timeFinish]
	timeTime := getTime(strTime)
	strDuration := getStrDuration(data[timeFinish+1:])
	strThread := getThread(data)

	if eventStartTime, ok := obj.processes[string(strProcess)]; ok {
		eventStopTime := timeTime
		eventDuration := eventStopTime.Sub(eventStartTime)

		if eventDuration < 0 {
			writer.Write(writeEvent("%s%s.%06d-%d,TIMEBACK\n",
				strProcess, eventStopTime.Format("06010215.log:04:05"),
				eventStopTime.Nanosecond()/1000,
				eventDuration.Abs().Milliseconds()))
		}
	}
	obj.processes[string(strProcess)] = timeTime

	thread := timeGapThread{string(strProcess), string(strThread)}
	if eventStartTime, ok := obj.threads[thread]; ok {

		eventStopTime := getStartTime(strTime, strDuration)
		eventDuration := eventStopTime.Sub(eventStartTime)

		writer.Write(writeEvent("%s%s.%06d-%d,TIMEGAP,OSThread=%s\n",
			strProcess, eventStopTime.Format("06010215.log:04:05"),
			eventStopTime.Nanosecond()/1000,
			eventDuration.Milliseconds(), strThread))
	}
	obj.threads[thread] = timeTime

	writer.Write(data)
	writer.Write([]byte("\n"))
}
