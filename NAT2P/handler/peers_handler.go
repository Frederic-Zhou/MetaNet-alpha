package handler

import (
	"encoding/json"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/network"
)

type Peers struct {
	Action string `json:"action"`
	Peers  []string
}

func (p Peers) Do(e *network.Event) (err error) {
	if e.GetDataType() != network.DataType_Text {
		return
	}

	err = json.Unmarshal(e.GetBody(), &p)
	if err != nil {
		return
	}

	switch p.Action {
	case "ask":

	case "relay":

	}

	return
}
