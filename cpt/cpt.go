package main

import (
	"fmt"
	"os"
)

var (
	version = "dev"
	//commit  = "none"
	date = "unknown"
)

type dataTransferType int

const (
	dataCopy dataTransferType = iota
	dataMove
)

func main() {
	var conf config

	if err := conf.init(os.Args); err != nil {
		switch err.(type) {
		case printVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case printUsage:

		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
		return
	}

	

	// cpt -m ./m1app 24022006.log ../case_2024.02.20/tj/m1app
	// cpt -m ./ 24022006.log //merope.dept07/csr/error/tj

}
