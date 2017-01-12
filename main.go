package main

import (
	"container/list"
	"log"
	"os"

	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type App struct {
	exts []string

	dirpath string
	command string
	dirList *list.List

	watcher *fsnotify.Watcher

	created bool
	writed  bool
	removed bool
	renamed bool
	chmod   bool

	debug bool

	chDirEvent  chan FileEvent
	chDone      chan bool
	chFileEvent chan FileEvent

	chSysSignal chan os.Signal
}

type FileEvent struct {
	action string
	file   string
}

func (app *App) checkExtension(path string) bool {
	if len(app.exts) == 0 {
		return true
	}
	fileExt := filepath.Ext(path)[1:]
	for _, v := range app.exts {
		if v == fileExt {
			return true
		}
	}
	return false
}

func (app *App) setWatchPath(path string) error {
	pathList, err := getDirectoryList(path)
	if err != nil {
		return err
	}

	for _, v := range pathList {
		if app.debug {
			log.Println(v)
		}
		app.dirList.PushBack(v)
		app.watcher.Add(v)
	}

	return nil
}

func (app *App) initWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	app.watcher = watcher
	return err
}

func main() {

	app, help, err := initSettings()
	if err != nil {
		log.Fatal(err)
	}

	if help {
		return
	}

	err = app.initWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer app.watcher.Close()

	go app.watchFiles()
	err = app.setWatchPath(app.dirpath)
	if err != nil {
		log.Fatal(err)
	}

exit:
	for {
		select {
		case e := <-app.chDirEvent:
			isDir, err := app.checkDir(&e)
			if err != nil {
				log.Println(err)
				break
			}
			if !isDir {
				break
			}

			switch e.action {
			case "created":
				app.setWatchPath(e.file)
			case "removed", "renamed":
				app.removePath(e.file)
			}
		case <-app.chDone:
			break exit
		}
	}
}
