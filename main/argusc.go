package main

import (
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"hardened_argus/argus"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	ENVUID        = "ARGUS_UID"
	ENVLABEL      = "ARGUS_LABEL"
	ENVSCRIPTROOT = "ARGUS_SCRIPT_ROOT"
)

const (
	PUBLIC_KEY = "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8HdKo6auCeQzld1Ojhtfp1RbYsg0sjQjLcR8wUcahx/Dg8d81tSn3sgkkVTLKZFX9y74HJujmZFIE+MlH1h2rw=="
)

func loadOrNewGuid() (string, error) {
	home, err := argus.HomeUnix()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, "argus.uid")
	_, err = os.Stat(path)
	if err == nil {
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		if len(buf) == 32 {
			return string(buf), nil
		}
		return "", errors.New("BAD GUID File")
	} else {
		if os.IsNotExist(err) {
			full := "" + time.Now().String()
			sh := md5.Sum([]byte(full))
			guid := hex.EncodeToString(sh[:])
			ioutil.WriteFile(path, []byte(guid), os.ModePerm)
			return guid, nil
		}
		return "", err
	}
}

func loadConfig() (*argus.ClientConfig, error) {
	conf := &argus.ClientConfig{}
	conf.UID = os.Getenv(ENVUID)
	if len(conf.UID) == 0 {
		return nil, errors.New("The value of environment variable named " + ENVUID + " not set")
	}
	conf.Label = os.Getenv(ENVLABEL)
	conf.GUID, _ = loadOrNewGuid()
	conf.ServerURL = "http://127.0.0.1:7890"
	conf.ScriptRoot = os.Getenv(ENVSCRIPTROOT)
	if len(conf.ScriptRoot) == 0 {
		home, _ := argus.HomeUnix()
		conf.ScriptRoot = filepath.Join(home, "argus_script")
		os.MkdirAll(conf.ScriptRoot, os.ModePerm)
	}
	return conf, nil
}

func main() {
	key, err := argus.StringToPublicKey(PUBLIC_KEY)
	if err != nil {
		fmt.Println(err)
		return
	}
	conf, err := loadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	guard, err := argus.NewArgusClient(conf, key.(*ecdsa.PublicKey))
	if err != nil {
		fmt.Println(err)
		return
	}
	guard.Run()
}
