package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var worker pathWalker

	worker.init()
	for _, path := range os.Args[1:] {
		worker.pathWalk(path)
	}
}

///////////////////////////////////////////////////////////////////////////////

type pathWalker struct {
	check lineChecker
}

func (obj *pathWalker) init() {
	obj.check.init()
}

func (obj *pathWalker) pathWalk(basePath string) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		obj.doProcess(basePath, path)

		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", basePath, err)
	}
}

func (obj *pathWalker) doProcess(basePath string, fileName string) {
	// path, err := os.Getwd()
	subFileName := path.Join(".", strings.Replace(fileName, basePath, "", 1))

	fileStream, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", fileName, err)
	}
	obj.check.processStream(subFileName, fileStream, os.Stdout)

}

///////////////////////////////////////////////////////////////////////////////

type lineChecker struct {
	checkFirstLine *regexp.Regexp
}

func (obj *lineChecker) init() {
	obj.checkFirstLine, _ = regexp.Compile(`^\d\d\:\d\d\.\d{6}\-\d+\,\w+\,`)
}

func (obj *lineChecker) processStream(sName string, sIn io.Reader, sOut io.Writer) {
	var str string

	scanner := bufio.NewScanner(sIn)

	if scanner.Scan() {
		str = scanner.Text()
		if str[:3] == "\ufeff" {
			str = str[3:]
		}
		fmt.Fprint(sOut, sName, ":", str)
	}

	for scanner.Scan() {
		str = scanner.Text()

		if obj.isFirstLine(str) {
			fmt.Fprint(sOut, "\n", sName, ":", str)
		} else {
			fmt.Fprint(sOut, "<line>", str)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}
}

func (obj *lineChecker) isFirstLine(data string) bool {
	return obj.checkFirstLine.MatchString(data)
}
