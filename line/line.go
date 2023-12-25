package main

import (
	"io"
	"os"
)

func main() {
	processFile(os.Stdin, os.Stdout)
}

func processFile(sIn io.Reader, sOut io.Writer) {

}
