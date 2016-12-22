package main

import (
	"os"
	"path/filepath"
	"strings"
)

func (app *App) removePath(path string) {
	for elem := app.dirList.Front(); elem != nil; elem.Next() {
		dirPath, ok := elem.Value.(string)
		if !ok {
			continue
		}
		if strings.Index(dirPath, path) == 0 {
			app.watcher.Remove(dirPath)
			app.dirList.Remove(elem)
		}
	}
	return
}

func (app *App) checkDir(e *FileEvent) (bool, error) {
	switch e.action {
	case "removed", "renamed":
		for elem := app.dirList.Front(); elem != nil; elem.Next() {
			if elem.Value == e.file {
				return true, nil
			}
		}
		return false, nil
	default:

		fInfo, err := os.Stat(e.file)
		if err != nil {
			return false, err
		}

		return fInfo.IsDir(), nil
	}

}

// searchExt searchs ext
// return true, if hit serching
func searchExt(fileName string, extList []string) bool {

	ext := filepath.Ext(fileName)

	if len(ext) <= 1 {
		return false
	}

	for _, e := range extList {
		if e == ext[1:] {
			return true
		}
	}

	return false
}

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
