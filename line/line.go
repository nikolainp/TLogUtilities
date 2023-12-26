package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main() {
	var check lineChecker

	check.init()
	check.processFile(os.Stdin, os.Stdout)
}

///////////////////////////////////////////////////////////////////////////////

type lineChecker struct {
	check *regexp.Regexp
}

func (obj *lineChecker) init() {
	obj.check, _ = regexp.Compile(`^\d\d\:\d\d\.\d{6}\-\d+\,\w+\,`)
}

func (obj *lineChecker) processFile(sIn io.Reader, sOut io.Writer) {
	var str string

	scanner := bufio.NewScanner(sIn)

	for scanner.Scan() {
		str = scanner.Text()
		if obj.isFirstLine(str) {
			fmt.Fprintln(sOut, str)
		} else {
			fmt.Fprintln(sOut, str)
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}
}

func (obj *lineChecker) isFirstLine(data string) bool {
	return obj.check.MatchString(data)
}
