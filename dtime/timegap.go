package main

import "io"

type timeGap struct {
}

func (obj *timeGap) lineProcessor(data []byte, writer io.Writer) {
	timeBegin, timeFinish := getStrTimePosition(data)
	if timeBegin == 0 && timeFinish == 0 {
		return
	}

	// strProcess := data[0:timeBegin]
	// strTime := data[timeBegin:timeFinish]

	writer.Write(data)
	writer.Write([]byte("\n"))
}
