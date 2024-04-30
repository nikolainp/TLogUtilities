package main

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type timeGapTread struct {
	process string
	tread   string
}

type timeGap struct {
	processes map[string]time.Time
	treads map[timeGapTread]time.Time
}

func (obj *timeGap) init() {
	obj.processes = make(map[string]time.Time)
	obj.treads = make(map[timeGapTread]time.Time)
}

func (obj *timeGap) lineProcessor(data []byte, writer io.Writer) {

	getTread := func(data []byte) []byte {
		treadMask := []byte(",OSTread=")
		treadPosition := bytes.Index(data, treadMask)
		commaPosition := bytes.Index(data[treadPosition+len(treadMask):], []byte(","))

		return data[treadPosition+len(treadMask) : treadPosition+len(treadMask)+commaPosition]
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
	strDuration := getStrDuration(data[timeFinish+1:])
	strTread := getTread(data)

	if eventStartTime,ok := obj.processes[string(strProcess)]; ok {
		eventStopTime := getTime(strTime)
		eventDuration := eventStopTime.Sub(eventStartTime)

		if eventDuration < 0 {
			writer.Write(writeEvent("%s%s.%06d-%d,TIMEBACK\n",
			strProcess, eventStopTime.Format("06010215.log:04:05"),
			eventStopTime.Nanosecond()/1000,
			eventDuration.Abs().Milliseconds()))
		}
	}
	obj.processes[string(strProcess)] = getTime(strTime)

	thread := timeGapTread{string(strProcess), string(strTread)}
	if eventStartTime, ok := obj.treads[thread]; ok {

		eventStopTime := getStartTime(strTime, strDuration)
		eventDuration := eventStopTime.Sub(eventStartTime)

		writer.Write(writeEvent("%s%s.%06d-%d,TIMEGAP,OSTread=%s\n",
			strProcess, eventStopTime.Format("06010215.log:04:05"),
			eventStopTime.Nanosecond()/1000,
			eventDuration.Milliseconds(), strTread))
	}
	obj.treads[thread] = getTime(strTime)

	writer.Write(data)
	writer.Write([]byte("\n"))
}
