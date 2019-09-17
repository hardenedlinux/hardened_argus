package argus

import (
	"time"
)

type machine struct {
	UID    string `json:"uid"`
	GUID   string `json:"guid"`
	Label  string `json:"label"`
	Active int64  `json:"last_active"`
}

type machineManger struct {
	machines map[string]*machine
}

func (o *machineManger) heartbeatNotify(req *heartbeat_request) {
	t := &machine{
		UID:    req.UID,
		GUID:   req.GUID,
		Label:  req.Label,
		Active: time.Now().Unix(),
	}
	o.machines[t.GUID] = t
}

func (o *machineManger) enumMachines() []*machine {
	list := make([]*machine, len(o.machines))
	i := 0
	for _, v := range o.machines {
		list[i] = v
		i++
	}
	return list
}
