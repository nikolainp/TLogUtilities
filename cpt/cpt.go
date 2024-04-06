package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	conf := getConfig(os.Args)

	var fp fileProcessor
	fp.init(conf.sourceFolder, conf.destinationFolder, conf.transferType)

	pathWalk(conf.sourceFolder, fp)
}

func getConfig(args []string) (conf config) {
	if err := conf.init(args); err != nil {
		switch err.(type) {
		case printVersion:
			fmt.Printf("Version: %s (%s)\n", version, date)
		case printUsage:

		default:
			fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		}
		os.Exit(0)
	}
	return conf
}

func pathWalk(basePath string, fp fileProcessor) {
	err := filepath.Walk(basePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		fp.doProcess(path)

		// if isCancel() {
		// 	return fmt.Errorf("process is cancel")
		// }

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path %q: %v\n", basePath, err)
	}
}

///////////////////////////////////////////////////////////////////////////////

type fileProcessor struct {
	source      string
	destination string

	transferType dataTransferType
}

func (obj *fileProcessor) init(source, destination string, transferType dataTransferType) {
	obj.source = source
	obj.destination = destination
	obj.transferType = transferType
}

func (obj *fileProcessor) doProcess(fileName string) {

	var err error

	getSubFilePath := func(path string) string {
		dir, file := filepath.Split(path)
		base := filepath.Base(dir)
		return filepath.Join(base, file)
	}

	subFilePath := getSubFilePath(fileName)
	destintion := filepath.Join(obj.destination, subFilePath)
	//err = os.MkdirAll(filepath.Dir(subFilePath), 0777)
	err = createDirectory(filepath.Dir(destintion))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	switch obj.transferType {
	case dataCopy:
		fmt.Printf("cp %s %s\n", fileName, destintion)
		err = copyFile(fileName, destintion)
	case dataMove:
		fmt.Printf("mv %s %s\n", fileName, destintion)
		err = moveFile(fileName, destintion)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

///////////////////////////////////////////////////////////////////////////////

func createDirectory(path string) error {

	// Check if the directory exists
	_, err := os.Stat(path)
	if err == nil {
		fmt.Printf("exists: %s\n", path)
		return nil
	}

	// If the directory does not exist, create its parent
	if os.IsNotExist(err) {
		err = createDirectory(filepath.Dir(path))
		if err != nil {
			return err
		}
		// Create the directory
		err = os.Mkdir(path, 0777)
		if err != nil {
			return err
		}
		fmt.Printf("created: %s\n", path)
	}
	return nil
}

func moveFile(src, dst string) error {
	return os.Rename(src, dst)
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)

	if nBytes != sourceFileStat.Size() {
		return fmt.Errorf("copy %d bytes from %d", nBytes, sourceFileStat.Size())
	}

	return err
}
