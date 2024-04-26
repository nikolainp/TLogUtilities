package main

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

type edgeType int

const (
	edgeNone edgeType = iota
	edgeStart
	edgeStop
	edgeActiveAll
	edgeActiveOnly
)

type lineFilter struct {
	timeBegin, timeFinish       time.Time
	endTimeBegin, endTimeFinish []byte
	edge                        edgeType

	filter func([]byte) bool
}

func (obj *lineFilter) init(start time.Time, stop time.Time, edge edgeType) {

	timeToFilter := func(tt time.Time) []byte {
		millisecond := tt.Nanosecond() / 1000

		return []byte(fmt.Sprintf("%02d%02d%02d%02d.log:%02d:%02d.%06d",
			tt.Year()-2000, tt.Month(), tt.Day(), tt.Hour(),
			tt.Minute(), tt.Second(), millisecond))
	}

	obj.timeBegin = start
	obj.timeFinish = stop
	obj.endTimeBegin = timeToFilter(start)
	obj.endTimeFinish = timeToFilter(stop)
	obj.edge = edge

	switch edge {
	case edgeStart:
		obj.filter = obj.isTrueLineByStart
	case edgeStop:
		obj.filter = obj.isTrueLineByStop
	case edgeActiveAll:
		obj.filter = obj.isTrueLineInvolve
	case edgeActiveOnly:
		obj.filter = obj.isTrueLineActive
	}
}

func (obj *lineFilter) LineProcessor(data []byte, writer io.Writer) {
	if obj.filter(data) {
		writer.Write(data)
		writer.Write([]byte("\n"))
	}
}

func (obj *lineFilter) isTrueLineByStart(data []byte) bool {
	strLineTime, strDuration := getStrTimeFromLine(data)
	if strLineTime == nil {
		return false
	}

	//eventStopMoment, _ := time.ParseDuration(string(strLineTime[19:]) + "us")
	eventStartTime := getStartTime(strLineTime, strDuration)

	if eventStartTime.Compare(obj.timeBegin) == -1 ||
		eventStartTime.Compare(obj.timeFinish) == 1 {
		return false
	}

	return true
}

func (obj *lineFilter) isTrueLineActive(data []byte) bool {
	strLineTime, strDuration := getStrTimeFromLine(data)
	if strLineTime == nil {
		return false
	}

	eventStartTime := getStartTime(strLineTime, strDuration)

	if eventStartTime.Compare(obj.timeBegin) == 1 ||
		bytes.Compare(strLineTime, obj.endTimeFinish) == -1 {
		return false
	}

	return true
}

func (obj *lineFilter) isTrueLineByStop(data []byte) bool {
	strLineTime, _ := getStrTimeFromLine(data)
	if strLineTime == nil {
		return false
	}
	if bytes.Compare(strLineTime, obj.endTimeBegin) == -1 ||
		bytes.Compare(strLineTime, obj.endTimeFinish) == 1 {
		return false
	}
	return true
}

func (obj *lineFilter) isTrueLineInvolve(data []byte) bool {
	return obj.isTrueLineByStop(data) ||
		obj.isTrueLineByStart(data) ||
		obj.isTrueLineActive(data)
}

///////////////////////////////////////////////////////////////////////////////

func getStrTimePosition(data []byte) (begin, finish int) {
	isNumber := func(data byte) bool {
		if data == '0' || data == '1' || data == '2' || data == '3' ||
			data == '4' || data == '5' || data == '6' || data == '7' ||
			data == '8' || data == '9' {
			return true
		}

		return false
	}

	logPositin := bytes.Index(data, []byte(".log:"))
	if logPositin < 8 || len(data) < logPositin+18 {
		return 0, 0
	}

	begin = logPositin - 8
	finish = logPositin + 17

	strTime := data[begin:finish]
	if !isNumber(strTime[0]) ||
		!isNumber(strTime[1]) ||
		!isNumber(strTime[2]) ||
		!isNumber(strTime[3]) ||
		!isNumber(strTime[4]) ||
		!isNumber(strTime[5]) ||
		!isNumber(strTime[6]) ||
		!isNumber(strTime[7]) {
		return 0, 0
	}
	if strTime[15] != ':' || strTime[18] != '.' {
		return 0, 0
	}
	if !isNumber(strTime[13]) ||
		!isNumber(strTime[14]) ||
		!isNumber(strTime[16]) ||
		!isNumber(strTime[17]) {
		return 0, 0
	}
	if !isNumber(strTime[19]) ||
		!isNumber(strTime[20]) ||
		!isNumber(strTime[21]) ||
		!isNumber(strTime[22]) ||
		!isNumber(strTime[23]) ||
		!isNumber(strTime[24]) {
		return 0, 0
	}

	return
}

func getStrDuration(data []byte) []byte {
	commaPosition := bytes.Index(data, []byte(","))
	if commaPosition == -1 {
		return nil
	}

	return data[:commaPosition]
}

func getStrTimeFromLine(data []byte) (time []byte, duration []byte) {

	timeBegin, timeFinish := getStrTimePosition(data)
	if timeBegin == 0 && timeFinish == 0 {
		return nil, nil
	}
	strTime := data[timeBegin:timeFinish]

	strDuration := getStrDuration(data[timeFinish+1:])
	if strDuration == nil {
		return nil, nil
	}

	return strTime, strDuration
}

func getTime(strLineTime []byte) time.Time {
	stopTime, _ := time.ParseInLocation("06010215.log:04:05", string(strLineTime), time.Local)

	return stopTime
}

func getStartTime(strLineTime []byte, strDuration []byte) time.Time {
	stopTime := getTime(strLineTime)

	duration, _ := time.ParseDuration(string(strDuration) + "us")
	startTime := stopTime.Add(-1 * duration)

	return startTime
}

///////////////////////////////////////////////////////////////////////////////
