package main

import (
	"encoding/json"
	"fmt"
	"hardened_argus/argus"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
)

const (
	productName = "argus"
)

//GetCurrentDirectory get app runing root path
func GetCurrentDirectory() string {
	execPath, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(execPath)
	dir := filepath.Dir(path)
	return dir
}

//ReadConfigFile read the config file
//if the name not contain '.',the file default is ".conf" file
//first read file at executeable file path /conf/
//second read file at /etc/$productname/
func ReadConfigFile(name string) ([]byte, error) {
	var path = GetCurrentDirectory() + string(filepath.Separator) + "conf" + string(filepath.Separator)
	path += name
	if !strings.Contains(name, ".") {
		path += ".conf"
	}
	b, err := ioutil.ReadFile(path)
	if err == nil {
		return b, err
	}
	path = "/etc/" + name
	if !strings.Contains(name, ".") {
		path += ".conf"
	}
	return ioutil.ReadFile(path)
}

func processExitSignal(srv *argus.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	srv.Shutdown()
	os.Exit(0)
}

func main() {
	conf := &argus.Config{}
	buf, err := ReadConfigFile(productName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(buf, conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = os.Stat(conf.LogPath)
	if os.IsNotExist(err) {
		os.MkdirAll(conf.LogPath, os.ModePerm)
	}

	srv, err := argus.NewServer(conf)
	if err != nil {
		fmt.Println(err)
		return
	}
	go processExitSignal(srv)
	err = srv.Run()
	srv.Shutdown()
	if err != nil {
		fmt.Println(err)
		return
	}
}
