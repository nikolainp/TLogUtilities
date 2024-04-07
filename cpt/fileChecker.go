package main

import (
	"fmt"
	"path/filepath"
)

type fileChecker struct {
	matches []string
}

func (obj *fileChecker) init(fileNames []string) error {
	obj.matches = append(obj.matches, fileNames...)

	for _, match := range obj.matches {

		if _, err := filepath.Match(match, ""); err != nil {
			return fmt.Errorf("wrong match format: %s", match)
		}
	}

	return nil
}

func (obj *fileChecker) isTrueFile(fileName string) bool {

	if len(obj.matches) == 0 {
		return true
	}

	check := func(name string) bool {
		for _, match := range obj.matches {

			if res, _ := filepath.Match(match, name); res {
				return true
			}
		}

		return false
	}

	if res := check(fileName); res {
		return true
	}

	if ext := filepath.Ext(fileName); ext != "" {
		if res := check(fileName[:len(fileName)-len(ext)]); res {
			return true
		}
	}

	return false
}

///////////////////////////////////////////////////////////////////////////////
