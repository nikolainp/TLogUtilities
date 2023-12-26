package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

func main() {
	processFile(os.Stdin, os.Stdout)
}

func processFile(sIn io.Reader, sOut io.Writer) {
	scanner := bufio.NewScanner(sIn)
	for scanner.Scan() {
		str := scanner.Text()
		fmt.Fprintln(sOut, str)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}
}

///////////////////////////////////////////////////////////////////////////////

type lineChecker struct {
	check *regexp.Regexp
}

func (obj *lineChecker) init() {
	obj.check, _ = regexp.Compile(`^\d\d\:\d\d\.\d{6}\-\d+\,\w+\,`)
}

func (obj lineChecker) isFirstLine(data string) bool {
	return obj.check.MatchString(data)
}
