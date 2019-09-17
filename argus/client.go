package argus

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//
type ClientConfig struct {
	UID        string `json:"uid"`
	GUID       string `json:"guid"`
	ScriptRoot string `json:"script_root"`
	Label      string `json:"label"`
	ServerURL  string `json:"server_address"`
}

//ArgusClient client object
type ArgusClient struct {
	conf      *ClientConfig
	publicKey *ecdsa.PublicKey
}

//HomeUnix() get home path
func HomeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func NewArgusClient(conf *ClientConfig, publicKey *ecdsa.PublicKey) (*ArgusClient, error) {
	if len(conf.UID) == 0 || len(conf.GUID) == 0 || len(conf.ScriptRoot) == 0 || len(conf.ServerURL) == 0 {
		return nil, errors.New("Invalid config")
	}
	obj := &ArgusClient{}
	obj.conf = conf
	obj.publicKey = publicKey
	return obj, nil
}

func (o *ArgusClient) doHeartbeat() ([]*taskResponse, error) {
	req := heartbeat_request{
		UID:   o.conf.UID,
		GUID:  o.conf.GUID,
		Label: o.conf.Label,
	}
	buf, _ := json.Marshal(req)
	rsp, err := http.Post(o.conf.ServerURL+"/client/active.do", "application/json", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode != http.StatusOK {
		return nil, errors.New(rsp.Status)
	}
	buf, err = ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	hbrsp := &heartbeat_response{}
	err = json.Unmarshal(buf, hbrsp)
	if err != nil {
		return nil, err
	}
	return hbrsp.Tasks, nil
}

func (o *ArgusClient) doPostExecute(taskid string, output string) error {
	values := map[string]string{}
	values["uid"] = o.conf.UID
	values["guid"] = o.conf.GUID
	values["taskid"] = taskid
	values["output"] = output
	buf, _ := json.Marshal(values)
	rsp, err := http.Post(o.conf.ServerURL+"/client/post.do", "application/json", bytes.NewReader(buf))
	if err != nil {
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		return errors.New(rsp.Status)
	}
	return nil
}

func (o *ArgusClient) urlFileName(urlText string) (string, error) {
	url, err := url.Parse(urlText)
	if err != nil {
		return "", err
	}
	ls := strings.Split(url.EscapedPath(), "/")
	return ls[len(ls)-1], nil
}

func (o *ArgusClient) downloadURL(url string) (string, error) {
	name, err := o.urlFileName(url)
	if err != nil {
		return "", err
	}
	path := filepath.Join(o.conf.ScriptRoot, name)
	_, err = os.Stat(path)
	if err == nil {
		return path, nil
	}
	rsp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	if rsp.StatusCode != http.StatusOK {
		return "", errors.New(url + " " + rsp.Status)
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, rsp.Body)
	return path, err
}

func (o *ArgusClient) verifyFile(script string, sigFile string) error {
	fscript, err := os.Open(script)
	if err != nil {
		return err
	}
	defer fscript.Close()

	chse := sha256.New()
	_, err = io.Copy(chse, fscript)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadFile(sigFile)
	if err != nil {
		return err
	}
	chs := chse.Sum(nil)
	err = Verify(chs, o.publicKey, string(buf))
	if err != nil {
		return err
	}
	return nil
}

func (o *ArgusClient) executeTask(task *taskResponse) ([]byte, error) {
	script, err := o.downloadURL(task.TaskURL)
	if err != nil {
		return nil, err
	}
	sig, err := o.downloadURL(task.TaskURL + ".sig")
	if err != nil {
		return nil, err
	}
	err = o.verifyFile(script, sig)
	if err != nil {
		return nil, err
	}
	os.Chmod(script, os.ModePerm)
	return exec.Command("/bin/bash", "-c", script).Output()
}

func (o *ArgusClient) Run() {
	for {
		rsp, err := o.doHeartbeat()
		if err != nil {
			fmt.Println(err)
		} else {
			for _, v := range rsp {
				out, err := o.executeTask(v)
				if err != nil {
					fmt.Println(err)
					o.doPostExecute(v.Taskid, err.Error())
				} else {
					o.doPostExecute(v.Taskid, string(out))
				}
			}
		}
		time.Sleep(time.Minute * 10)
	}
}
