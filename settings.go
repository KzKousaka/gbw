package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"strings"
)

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
