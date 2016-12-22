package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"syscall"

	"time"

	"strings"

	shellwords "github.com/mattn/go-shellwords"
	"github.com/mgutz/ansi"
)

const (
	STDOUT_COLOR = "green"
	STDERR_COLOR = "red"
)

const (
	C_END = iota
	C_WAMP
	C_WPIPE
	C_SCOLON
)

func (app *App) commandRunner() {
	var commandPerser *shellwords.Parser
	if app.command == "" {
		return
	}

	commandPerser = shellwords.NewParser()
	commandPerser.ParseEnv = true
	cArgs, err := commandPerser.Parse(app.command)
	if err != nil {
		log.Println(err)
		return
	}

	r := InitRunner()
	var last = time.Now().Add(-3 * time.Second)

	for {
		select {
		case e := <-app.chFileEvent:
			if app.debug {
				log.Println(e.action, ":", e.file)
			}

			if app.command == "" {
				break
			}

			if !app.checkExtension(e.file) {
				break
			}

			now := time.Now()
			if now.Sub(last) < 3*time.Second {
				break
			}

			last = now
			go r.pipeSplite(cArgs)
		}
	}
}

type Runner struct {
	chKill chan bool
	chExit chan error
	chJob  chan bool
}

func InitRunner() *Runner {

	return &Runner{
		chKill: make(chan bool),
		chExit: make(chan error),
		chJob:  make(chan bool, 1),
	}

}

func (r *Runner) pipeSplite(args []string) {
	if len(args) == 0 {
		return
	}

	if len(r.chJob) > 0 {
		r.chKill <- true
	}
	r.chJob <- true
	defer func() {
		<-r.chJob
		r.chKill = make(chan bool)
		r.chExit = make(chan error)
	}()

	time.Sleep(time.Second)

	cl := 0
	cr := 0

	for {
		cType := C_END
		skip := false

		for cr = cl; cr < len(args) && cType == C_END; cr++ {
			switch args[cr] {
			case "&&":
				cType = C_WAMP
			case "||":
				cType = C_WPIPE
			case ";":
				cType = C_SCOLON
			}
		}

		if skip {
			cl = cr + 1
			skip = false
			continue
		}

		var cmd *exec.Cmd
		if cr-cl == 1 {
			fmt.Printf("\x1b[1mrunning $ %s\x1b[0m\n", args[cl])
			cmd = exec.Command(args[cl])
		} else {
			fmt.Printf("\x1b[1mrunning $ %s %s\x1b[0m\n", args[cl], strings.Join(args[cl+1:cr-1], " "))
			cmd = exec.Command(args[cl], args[cl+1:cr-1]...)
		}

		exitCode, err, kill := r.command(cmd)
		if err != nil {
			log.Println(err)
		}
		if kill {
			break
		}

		switch cType {
		case C_WAMP:
			skip = exitCode != 0
		case C_WPIPE:
			skip = exitCode == 0
		case C_SCOLON:
			skip = false
		}

		cl = cr

		if cl >= len(args) {
			break
		}
	}
}

func (r *Runner) command(cmd *exec.Cmd) (exitCode int, err error, kill bool) {
	outReader, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	errReader, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	var bufout, buferr bytes.Buffer
	outReader2 := io.TeeReader(outReader, &bufout)
	errReader2 := io.TeeReader(errReader, &buferr)

	if err = cmd.Start(); err != nil {
		return
	}

	go r.printOutputWithHeader(STDOUT_COLOR, outReader2)
	go r.printOutputWithHeader(STDERR_COLOR, errReader2)

	go func() {
		err := cmd.Wait()
		r.chExit <- err
	}()

	killed := false
exit:
	for {
		select {
		case err := <-r.chExit:
			if err != nil {
				if err2, ok := err.(*exec.ExitError); ok {
					if s, ok := err2.Sys().(syscall.WaitStatus); ok {
						err = nil
						exitCode = s.ExitStatus()
					}
				}
			}
			break exit
		case <-r.chKill:
			if !killed {
				cmd.Process.Kill()
				killed = true
			}
		}
	}
	return
}

func (r *Runner) printOutputWithHeader(color string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Printf("%s\n", ansi.Color(scanner.Text(), color))
	}
}
