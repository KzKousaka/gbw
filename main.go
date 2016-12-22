package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"os/exec"
	"path/filepath"

	shellwords "github.com/mattn/go-shellwords"

	"github.com/go-fsnotify/fsnotify"
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
}

type FileEvent struct {
	action string
	file   string
}

const helpText = `
  -command string
        run command
  -dir string
        watching directory dirpath ` + "\x1b[2m(default \"./\")\x1b[0m" + `
  -ext string
        extension list ` + "\x1b[2m(ex \"png,gif\")\x1b[0m" + `
  -help this view

  -debug output debugging log

  ` + "\x1b[1m[event option]\x1b[0m" + `
  created:
    -c    enabled ` + "\x1b[2m(default)\x1b[0m" + `
    -nc   disabled
  writed:
    -w    enabled ` + "\x1b[2m(default)\x1b[0m" + `
    -nw   disabled
  removed:
    -r    enabled ` + "\x1b[2m(default)\x1b[0m" + `
    -nr   disabled
  renamed:
    -n    enabled ` + "\x1b[2m(default)\x1b[0m" + `
    -nn   disabled
  change permission:
    -p    enabled
    -np   disabled ` + "\x1b[2m(default)\x1b[0m" + `
`

func initSettings() (*App, bool, error) {
	var exts string
	app := &App{
		dirList:     list.New(),
		chDirEvent:  make(chan FileEvent, 100),
		chFileEvent: make(chan FileEvent, 100),
		chDone:      make(chan bool),
	}
	flag.StringVar(&app.dirpath, "dir", "./", "watching directory dirpath")
	flag.StringVar(&app.command, "command", "", "Run command")
	flag.StringVar(&exts, "ext", "", "extension list  ex:\"png,gif\"")

	var (
		ncreated bool
		nwrited  bool
		nremoved bool
		nrenamed bool
		nchmod   bool
		help     bool
	)
	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&app.debug, "debug", false, "")

	flag.BoolVar(&app.created, "c", false, "enabled event for created")
	flag.BoolVar(&ncreated, "nc", false, "disabled event for created")
	flag.BoolVar(&app.writed, "w", false, "enabled event for writed")
	flag.BoolVar(&nwrited, "nw", false, "disabled event for writed")
	flag.BoolVar(&app.removed, "r", false, "enabled event for removed")
	flag.BoolVar(&nremoved, "nr", false, "disabled event for removed")
	flag.BoolVar(&app.renamed, "n", false, "enabled event for renamed")
	flag.BoolVar(&nrenamed, "nn", false, "disabled event for renamed")
	flag.BoolVar(&app.chmod, "p", false, "enabled event for changed permisson")
	flag.BoolVar(&nchmod, "np", false, "disabled event for changed permisson")

	flag.Parse()

	if help {
		fmt.Println(helpText)
		return nil, true, nil
	}

	if app.created && ncreated {
		return nil, false, fmt.Errorf("-c , -nc can not be specified at the same time")
	}
	if app.writed && nwrited {
		return nil, false, fmt.Errorf("-w , -nw can not be specified at the same time")
	}
	if app.removed && nremoved {
		return nil, false, fmt.Errorf("-r , -nr can not be specified at the same time")
	}
	if app.renamed && nrenamed {
		return nil, false, fmt.Errorf("-n , -nn can not be specified at the same time")
	}
	if app.chmod && nchmod {
		return nil, false, fmt.Errorf("-p , -np can not be specified at the same time")
	}

	if !ncreated {
		app.created = true
	}
	if !nwrited {
		app.writed = true
	}
	if !nremoved {
		app.removed = true
	}
	if !nrenamed {
		app.renamed = true
	}
	if !app.chmod {
		app.chmod = false
	}

	if app.debug {
		log.SetFlags(log.Lshortfile | log.Ltime)
	}

	app.setExtList(exts)

	return app, false, nil
}

// extList returns the extension list
func (app *App) setExtList(exts string) []string {

	app.exts = []string{}
	if exts != "" {
		app.exts = strings.Split(exts, ",")
	}

	return app.exts
}

func (app *App) watchFiles() {
	go app.commandRunner()
	for {
		select {
		case event := <-app.watcher.Events:
			if len(app.exts) != 0 && !searchExt(event.Name, app.exts) {
				break
			}

			cmdRun := false

			var fEvent FileEvent

			switch {
			case event.Op&fsnotify.Write == fsnotify.Write && app.writed:
				fEvent.action = "writed"
				cmdRun = true

			case event.Op&fsnotify.Create == fsnotify.Create && app.created:
				fEvent.action = "created"
				cmdRun = true

			case event.Op&fsnotify.Remove == fsnotify.Remove && app.removed:
				fEvent.action = "removed"
				cmdRun = true

			case event.Op&fsnotify.Rename == fsnotify.Rename && app.renamed:
				fEvent.action = "renamed"
				cmdRun = true

			case event.Op&fsnotify.Chmod == fsnotify.Chmod && app.chmod:
				fEvent.action = "chmod"
				cmdRun = true

			}

			if cmdRun {
				fEvent.file = event.Name
				app.chDirEvent <- fEvent
				app.chFileEvent <- fEvent
			}

		case err := <-app.watcher.Errors:
			log.Println(err)
			app.chDone <- true
		}
	}
}

func (app *App) commandRunner() {
	var commandPerser *shellwords.Parser
	if app.command != "" {
		commandPerser = shellwords.NewParser()
		commandPerser.ParseEnv = true
	}

	job := make(chan bool, 1)

	for {
		select {
		case e := <-app.chFileEvent:
			if app.debug {
				log.Println(e.action, e.file)
			}

			if app.command == "" {
				break
			}

			if !app.checkExtension(e.file) {
				break
			}

			args, err := commandPerser.Parse(app.command)
			if err != nil {
				break
			}

			go func(args []string, e FileEvent) {

				select {
				case job <- true:
				default:
					log.Println("cancel :", e.file)
					return
				}

				if len(args) == 0 {
					return
				}

				var cmd *exec.Cmd
				if len(args) == 1 {
					cmd = exec.Command(args[0])
				} else {
					cmd = exec.Command(args[0], args[1:]...)
				}

				out, err := cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Println(string(out))
				}

				select {
				case <-job:
				default:
					return
				}
			}(args, e)
		}
	}
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
