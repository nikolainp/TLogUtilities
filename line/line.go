package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	processFile(os.Stdin, os.Stdout)
}

func processFile(sIn io.Reader, sOut io.Writer) {
	scanner := bufio.NewScanner(sIn)
	for scanner.Scan() {
		fmt.Fprintln(sOut, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}
}
