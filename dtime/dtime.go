package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

func main() {
	var conf config

	conf.init(os.Args)

	switch conf.getOperation() {
	case operationFilterByTyme:
	// filter by time: start finish edgeType

	case operationTimeGapBack:
		// add TIMEGAP TIMEBACK events
	}

	// operations with time
}

///////////////////////////////////////////////////////////////////////////////

///////////////////////////////////////////////////////////////////////////////

type edgeType int

const (
	edgeStart edgeType = iota
	edgeStop  
	edgeActiveAll
	edgeActiveOnly 
)

type lineFilter struct {
	timeBegin, timeFinish       time.Time
	endTimeBegin, endTimeFinish []byte
	edge                        edgeType
}

func (obj *lineFilter) init(start time.Time, stop time.Time, edge edgeType) {

	timeToFilter := func(tt time.Time) []byte {
		return []byte(fmt.Sprintf("%02d%02d%02d%02d.log:%02d:%02d.%06d",
			tt.Year()-2000, tt.Month(), tt.Day(), tt.Hour(),
			tt.Minute(), tt.Second(), tt.Nanosecond()))
	}

	obj.timeBegin = start
	obj.timeFinish = stop
	obj.endTimeBegin = timeToFilter(start)
	obj.endTimeFinish = timeToFilter(stop)
	obj.edge = edge
}

func (obj *lineFilter) isTrueLineByStart(data []byte) bool {
	strLineTime, strDuration := getStrTimeFromLine(data)
	if strLineTime == nil {
		return false
	}

	eventStopTime, _ := time.ParseInLocation("06010215.log:04:05", string(strLineTime), time.Local)
	//eventStopMoment, _ := time.ParseDuration(string(strLineTime[19:]) + "us")
	duration, _ := time.ParseDuration(string(strDuration) + "us")
	eventStartTime := eventStopTime.Add(-1 * duration)

	if eventStartTime.Compare(obj.timeBegin) == -1 ||
		eventStartTime.Compare(obj.timeFinish) == 1 {
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

///////////////////////////////////////////////////////////////////////////////

func getStrTimeFromLine(data []byte) (time []byte, duration []byte) {

	isNumber := func(data byte) bool {
		if data == '0' || data == '1' || data == '2' || data == '3' ||
			data == '4' || data == '5' || data == '6' || data == '7' ||
			data == '8' || data == '9' {
			return true
		}

		return false
	}

	logPositin := bytes.Index(data, []byte(".log:"))
	if logPositin < 8 || len(data) < logPositin+17 {
		return nil, nil
	}

	strTime := data[logPositin-8 : logPositin+17]
	if !isNumber(strTime[0]) ||
		!isNumber(strTime[1]) ||
		!isNumber(strTime[2]) ||
		!isNumber(strTime[3]) ||
		!isNumber(strTime[4]) ||
		!isNumber(strTime[5]) ||
		!isNumber(strTime[6]) ||
		!isNumber(strTime[7]) {
		return nil, nil
	}
	if strTime[15] != ':' || strTime[18] != '.' {
		return nil, nil
	}
	if !isNumber(strTime[13]) ||
		!isNumber(strTime[14]) ||
		!isNumber(strTime[16]) ||
		!isNumber(strTime[17]) {
		return nil, nil
	}
	if !isNumber(strTime[19]) ||
		!isNumber(strTime[20]) ||
		!isNumber(strTime[21]) ||
		!isNumber(strTime[22]) ||
		!isNumber(strTime[23]) ||
		!isNumber(strTime[24]) {
		return nil, nil
	}

	commaPosition := bytes.Index(data[logPositin+18:], []byte(","))
	if commaPosition == -1 {
		return nil, nil
	}
	strDuration := data[logPositin+18 : logPositin+18+commaPosition]

	return strTime, strDuration
}

///////////////////////////////////////////////////////////////////////////////
