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
	treads map[timeGapTread]time.Time
}

func (obj *timeGap) init() {
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
