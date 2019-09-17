package argus

import (
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//Config server config
type Config struct {
	LogPath string `json:"logdir"`
	BDPath  string `json:"dbdir"`
	Address string `json:"address"`
}

//Server hlkeep server
type Server struct {
	mm      *machineManger
	tasks   map[string][]*task
	taskMgr *taskManger
	conf    *Config
}

//client interface
func (o *Server) clientHeartbeatRequestHandler(w http.ResponseWriter, r *http.Request) {
	req := &heartbeat_request{
		UID:   r.FormValue("uid"),
		GUID:  r.FormValue("guid"),
		Label: r.FormValue("label"),
	}
	if len(req.UID) == 0 || len(req.GUID) != 32 {
		http.NotFound(w, r)
		return
	}
	o.mm.heartbeatNotify(req)
	tasks := []*taskResponse{}
	v, bfound := o.tasks[req.GUID]
	if bfound {
		for _, task := range v {
			if !o.taskMgr.isTaskInHistory(task.Taskid, req.GUID) {
				tasks = append(tasks, &taskResponse{
					Taskid:  task.Taskid,
					TaskURL: task.TaskURL,
				})
			}
		}
	}
	rsp := heartbeat_response{
		Timestamp: time.Now().Unix(),
		Tasks:     tasks,
	}
	buf, _ := json.Marshal(rsp)
	w.Write(buf)
}

func (o *Server) clientPostExecuteHandler(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	guid := r.FormValue("guid")
	taskid := r.FormValue("taskid")
	output := r.FormValue("output")
	name := uid + "_" + guid + "_" + taskid + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	if strings.Contains(name, "\\") || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(o.conf.LogPath, name)
	f, err := os.Create(path)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	f.Write([]byte(output))
	f.Close()
	w.Write([]byte("ok"))
	o.taskMgr.addPostExecute(taskid, guid)
	return
}

//machine manger interface
func (o *Server) mangerEnumManchineHandler(w http.ResponseWriter, r *http.Request) {
	list := o.mm.enumMachines()
	buf, err := json.Marshal(list)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(buf)
}

//task manger interface
func (o *Server) mangerEnumTaskHandler(w http.ResponseWriter, r *http.Request) {
	ts, ps := o.taskMgr.enumTask()
	rsp := &taskEnum_response{
		Tasks:  ts,
		Policy: ps,
	}
	buf, err := json.Marshal(rsp)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(buf)
}

func (o *Server) mangerAddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	url := r.FormValue("url")
	policy := r.FormValue("guids")
	period := r.FormValue("period")
	number := int64(0)
	if len(period) != 0 {
		number, err = strconv.ParseInt(period, 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	if len(url) == 0 {
		http.NotFound(w, r)
		return
	}
	guids := strings.Split(policy, "|")
	for _, v := range guids {
		if len(v) != 32 {
			http.NotFound(w, r)
			return
		}
	}
	t := &task{
		TaskURL: url,
		Period:  number,
	}
	err = o.taskMgr.addTask(t)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	err = o.taskMgr.updatePolicy(t.Taskid, guids)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	buf, _ := json.Marshal(t)
	w.Write(buf)
	o.reloadTask()
	return
}

func (o *Server) mangerDeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskid := r.FormValue("taskid")
	err := o.taskMgr.delTask(taskid)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("ok"))
	o.reloadTask()
	return
}

func (s *Server) parseForm(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Content-Type") == "application/json" {
		r.Form = make(url.Values)
		r.PostForm = make(url.Values)
		buf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return false
		}
		if buf == nil || len(buf) == 0 {
			http.Error(w, err.Error(), 500)
			return false
		}
		ret := make(map[string]string)
		err = json.Unmarshal(buf, &ret)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return false
		}
		for i, v := range ret {
			r.Form.Add(i, v)
			r.PostForm.Add(i, v)
		}
	} else {
		r.ParseForm()
	}
	return true
}

func (o *Server) reloadTask() {
	tasks := map[string][]*task{}
	ts, gs := o.taskMgr.enumTask()
	for i, t := range ts {
		for _, p := range gs[i] {
			ls := tasks[p]
			if ls == nil {
				ls = []*task{}
			}
			ls = append(ls, t)
			tasks[p] = ls
		}
	}
	o.tasks = tasks
}

//ServeHTTP http request handler
func (o *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !o.parseForm(w, r) {
		return
	}
	if r.URL.EscapedPath() == "/client/active.do" {
		o.clientHeartbeatRequestHandler(w, r)
		return
	}
	if r.URL.EscapedPath() == "/client/post.do" {
		o.clientPostExecuteHandler(w, r)
		return
	}
	if r.URL.EscapedPath() == "/manger/addtask.do" {
		o.mangerAddTaskHandler(w, r)
		return
	}
	if r.URL.EscapedPath() == "/manger/enumtask.do" {
		o.mangerEnumTaskHandler(w, r)
		return
	}
	if r.URL.EscapedPath() == "/manger/deltask.do" {
		o.mangerDeleteTaskHandler(w, r)
		return
	}
	if r.URL.EscapedPath() == "/manger/enummachine.do" {
		o.mangerEnumManchineHandler(w, r)
		return
	}
	http.NotFound(w, r)
}

//Shutdown to free resource
func (o *Server) Shutdown() {
	o.taskMgr.db.Close()
}

func (o *Server) Run() error {
	server := http.Server{}
	server.Addr = o.conf.Address
	server.Handler = o
	server.ReadTimeout = 2 * time.Minute
	server.WriteTimeout = 2 * time.Minute
	return server.ListenAndServe()
}

func NewServer(conf *Config) (*Server, error) {
	db, err := leveldb.OpenFile(conf.BDPath, nil)
	if err != nil {
		return nil, err
	}
	s := &Server{
		taskMgr: &taskManger{db: db},
		mm:      &machineManger{machines: make(map[string]*machine)},
		conf:    conf,
		tasks:   make(map[string][]*task),
	}
	s.reloadTask()
	return s, nil
}
