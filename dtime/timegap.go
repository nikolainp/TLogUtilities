package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"
)

type timeGapThread struct {
	process string
	thread  string
}

type timeGap struct {
	processes map[string]time.Time
	threads   map[timeGapThread]*timeLevel
}

func (obj *timeGap) init() {
	obj.processes = make(map[string]time.Time)
	obj.threads = make(map[timeGapThread]*timeLevel)
}

func (obj *timeGap) lineProcessor(data []byte, writer io.Writer) {

	getLevel := func(data []byte) int {
		levelPosition := bytes.Index(data, []byte(","))
		if levelPosition == -1 {
			return 0
		}
		commaPosition := bytes.Index(data[levelPosition+1:], []byte(","))
		if commaPosition == -1 {
			return 0
		}
		dataLevel := data[levelPosition+1 : levelPosition+1+commaPosition]
		if level, err := strconv.Atoi(string(dataLevel)); err == nil {
			return level
		}

		return 0
	}
	getThread := func(data []byte) []byte {
		threadMask := []byte(",OSThread=")
		threadPosition := bytes.Index(data, threadMask)
		if threadPosition == -1 {
			return []byte{}
		}
		commaPosition := bytes.Index(data[threadPosition+len(threadMask):], []byte(","))
		if commaPosition == -1 {
			return data[threadPosition+len(threadMask):]
		}

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
	newEventTime := getTime(strTime)
	strDuration, durationFinish := getStrDuration(data[timeFinish+1:])
	level := getLevel(data[timeFinish+1+durationFinish+1:])
	strThread := getThread(data)

	if eventStartTime, ok := obj.processes[string(strProcess)]; ok {
		eventStopTime := newEventTime
		eventDuration := eventStopTime.Sub(eventStartTime)

		if eventDuration < 0 {
			writer.Write(writeEvent("%s%s.%06d-%d,TIMEBACK\n",
				strProcess, eventStopTime.Format("06010215.log:04:05"),
				eventStopTime.Nanosecond()/1000,
				eventDuration.Abs().Microseconds()))
		}
	}
	obj.processes[string(strProcess)] = newEventTime

	thread := timeGapThread{string(strProcess), string(strThread)}
	if levelTime, ok := obj.threads[thread]; ok {

		eventStartTime := levelTime.add(level, newEventTime)
		if !eventStartTime.IsZero() {
			eventStopTime := getStartTime(strTime, strDuration)
			eventDuration := eventStopTime.Sub(eventStartTime)

			writer.Write(writeEvent("%s%s.%06d-%d,TIMEGAP,%d,OSThread=%s,\n",
				strProcess, eventStopTime.Format("06010215.log:04:05"),
				eventStopTime.Nanosecond()/1000,
				eventDuration.Microseconds(),
				level, strThread))
		}

	} else {
		obj.threads[thread] = newLevelTime(level, newEventTime)
	}

	writer.Write(data)
	writer.Write([]byte("\n"))
}

///////////////////////////////////////////////////////////////////////////////

type timeLevel struct {
	level     int
	levelTime []time.Time
}

func newLevelTime(level int, levelTime time.Time) *timeLevel {
	obj := new(timeLevel)
	obj.levelTime = make([]time.Time, 1)
	obj.add(level, levelTime)
	return obj
}
func (obj *timeLevel) add(level int, levelTime time.Time) time.Time {

	if level <= obj.level {
		obj.level = level
		obj.levelTime = obj.levelTime[:level+1]
		oldTime := obj.levelTime[level]
		obj.levelTime[level] = levelTime
		return oldTime
	}

	for i := obj.level; i < level; i++ {
		obj.levelTime = append(obj.levelTime, time.Time{})
	}
	obj.level = level
	obj.levelTime[level] = levelTime

	return time.Time{}
}
