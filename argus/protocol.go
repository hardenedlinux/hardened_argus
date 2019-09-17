package argus

type heartbeat_request struct {
	UID   string `json:"uid"`
	GUID  string `json:"guid"`
	Label string `json:"label"`
}

type taskResponse struct {
	Taskid  string `json:"taskid"`
	TaskURL string `json:"taskurl"`
}

type heartbeat_response struct {
	Timestamp int64           `json:"timestamp"`
	Tasks     []*taskResponse `json:"tasks"`
}

type taskEnum_response struct {
	Tasks  []*task    `json:"tasks"`
	Policy [][]string `json:"policy"`
}
