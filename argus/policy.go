package argus

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"strconv"
	"time"
)

type task struct {
	Taskid  string `json:"taskid"`  //任务id，addTask的时候生成
	TaskURL string `json:"taskurl"` //任务URL
	Period  int64  `json:"period"`  //触发周期
}

type taskManger struct {
	db *leveldb.DB
}

func (o *taskManger) updateTaskList(list []string) error {
	buf, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = o.db.Put([]byte("argus/tasklist"), buf, nil)
	if err != nil {
		return err
	}
	return err
}

func (o *taskManger) getTaskList() ([]string, error) {
	buf, err := o.db.Get([]byte("argus/tasklist"), nil)
	if err != nil {
		return []string{}, err
	}
	list := []string{}
	err = json.Unmarshal(buf, &list)
	if err != nil {
		return []string{}, err
	}
	return list, err
}

func (o *taskManger) newTaskID(seed string) string {
	full := seed + time.Now().String()
	sh := md5.Sum([]byte(full))
	return hex.EncodeToString(sh[:])
}

func (o *taskManger) addTask(task *task) error {
	task.Taskid = o.newTaskID(task.TaskURL)
	buf, err := json.Marshal(task)
	if err != nil {
		return err
	}
	key := "argus/tasks/" + task.Taskid
	err = o.db.Put([]byte(key), buf, nil)
	if err != nil {
		return err
	}
	list, _ := o.getTaskList()
	list = append(list, task.Taskid)
	return o.updateTaskList(list) //需要回滚？
}

func (o *taskManger) delTask(id string) error {
	key := "argus/tasks/" + id
	list, _ := o.getTaskList()
	newList := []string{}
	for _, v := range list {
		if v != id {
			newList = append(newList, v)
		}
	}
	o.db.Delete([]byte(key), nil)
	key = "argus/policy/" + id
	o.db.Delete([]byte(key), nil)
	prefix := "argus/tasks/history/" + id + "/"
	iter := o.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	for iter.Next() {
		o.db.Delete(iter.Key(), nil)
	}
	iter.Release()
	return o.updateTaskList(newList)
}

func (o *taskManger) getTask(id string) (*task, error) {
	key := "argus/tasks/" + id
	buf, err := o.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	t := &task{}
	err = json.Unmarshal(buf, t)
	if err != nil {
		return nil, err
	}
	return t, err
}

func (o *taskManger) enumTask() ([]*task, [][]string) {
	list, _ := o.getTaskList()
	tasks := []*task{}
	guids := [][]string{}
	fmt.Println(list)
	for _, v := range list {
		t, err := o.getTask(v)
		if t != nil {
			gs, _ := o.getPolicy(v)
			tasks = append(tasks, t)
			guids = append(guids, gs)
		} else {
			fmt.Println(err)
		}

	}
	return tasks, guids
}

func (o *taskManger) updatePolicy(taskid string, list []string) error {
	key := "argus/tasks/policy/" + taskid
	buf, err := json.Marshal(list)
	if err != nil {
		return err
	}
	err = o.db.Put([]byte(key), buf, nil)
	if err != nil {
		return err
	}
	return err
}

func (o *taskManger) getPolicy(taskid string) ([]string, error) {
	key := "argus/tasks/policy/" + taskid
	buf, err := o.db.Get([]byte(key), nil)
	if err != nil {
		return []string{}, err
	}
	list := []string{}
	err = json.Unmarshal(buf, &list)
	if err != nil {
		return []string{}, err
	}
	return list, err
}

func (o *taskManger) addPostExecute(taskid string, guid string) error {
	key := "argus/tasks/history/" + taskid + "/" + guid
	t, err := o.getTask(taskid)
	if err != nil {
		return err
	}
	expire := int64(0)
	if t.Period != 0 {
		expire = time.Now().Add(time.Minute * time.Duration(t.Period)).Unix()
	}
	text := strconv.FormatInt(expire, 10)
	return o.db.Put([]byte(key), []byte(text), nil)

}

func (o *taskManger) isTaskInHistory(taskid string, guid string) bool {
	key := "argus/tasks/history/" + taskid + "/" + guid
	buf, err := o.db.Get([]byte(key), nil)
	if err != nil {
		return false
	}
	value, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return false
	}
	if value == 0 {
		return true
	}
	expire := time.Unix(value, 0)
	if time.Now().After(expire) {
		return false
	}
	return true
}
