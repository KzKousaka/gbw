package main

import (
	"os"
	"path/filepath"
)

func getDirectoryList(path string) ([]string, error) {
	result := []string{}

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			result = append(result, path)
		}
		return nil
	})

	return result, nil
}
