package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type dummyFileManager struct {
	root string
	list []string
}

var _dfm *dummyFileManager

func (d *dummyFileManager) clean() {
	os.RemoveAll(d.root)
}

func makeDummyFile(path string) error {
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fmt.Fprint(fp, path)
	fp.Close()

	return nil
}

func makeDummyFiles() (*dummyFileManager, error) {
	path, _ := filepath.Abs(".")
	log.Println("current dir : ", path)
	os.MkdirAll("./test-dir/dir1", 0777)
	os.MkdirAll("./test-dir/dir2", 0777)
	os.MkdirAll("./test-dir/dir2/dir3", 0777)

	result := &dummyFileManager{
		root: "./test-dir",
		list: []string{
			"./test-dir/test1.aaa",
			"./test-dir/dir2/test2.bbb",
			"./test-dir/dir2/dir3/test3.ccc",
		},
	}

	for _, v := range result.list {
		if err := makeDummyFile(v); err != nil {
			return nil, err
		}
	}

	return result, nil

}

func TestMain(m *testing.M) {
	exitCode := 0
	defer os.Exit(exitCode)

	log.SetFlags(log.Ltime | log.Lshortfile)

	var err error
	_dfm, err = makeDummyFiles()
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer _dfm.clean()

	exitCode = m.Run()
}

func Test_getDirectorys(t *testing.T) {

	path := fmt.Sprintf("%v", _dfm.root)
	list, _ := getDirectoryList(path)
	if len(list) != 4 {
		log.Fatalf("Wrong value")
	}

	path = fmt.Sprintf("%v/", _dfm.root)
	list, _ = getDirectoryList(path)
}
