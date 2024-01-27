package main

import (
	"time"
)

type operationType int

const (
	operationNone operationType = iota
	operationFilterByTyme
	operationTimeGapBack
)

type config struct {
	programName string
	operation   operationType

	filterBeginTime  time.Time
	filterFinishTime time.Time
	filterEdge       edgeType
}

func (obj *config) init(args []string) (err error) {

	obj.programName = args[0]
	obj.operation = operationNone

	getFilterTime := func(data string) (time.Time, error) {
		time, err := time.ParseInLocation("2006.01.02_15:04:05", string(data), time.Local)
		return time, err
	}

	switch len(args) {
	case 1:
		obj.operation = operationTimeGapBack
	case 3:
		obj.operation = operationFilterByTyme

		obj.filterBeginTime, err = getFilterTime(args[1])
		if err != nil {
			return err
		}
		obj.filterFinishTime, err = getFilterTime(args[2])
		if err != nil {
			return err
		}
		obj.filterEdge = edgeStop
	case 4:
		obj.operation = operationFilterByTyme

		obj.filterBeginTime, err = getFilterTime(args[1])
		if err != nil {
			return err
		}
		obj.filterFinishTime, err = getFilterTime(args[2])
		if err != nil {
			return err
		}
		switch args[3] {
		case "start":
			obj.filterEdge = edgeStart
		case "stop":
			obj.filterEdge = edgeStop
		default:
			obj.filterEdge = edgeStart
		}
	}

	return nil
}

func (obj *config) getOperation() operationType {
	return obj.operation
}
