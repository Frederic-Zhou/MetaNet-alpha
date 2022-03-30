package handler

import (
	"bytes"
	"encoding/json"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/network"
)

type Peers struct {
	Action string `json:"action"`
	Peers  []string
}

var peers []string

func (p Peers) Do(e *network.Event) (err error) {

	err = json.Unmarshal(e.GetBody(), &p)
	if err != nil {
		return
	}

	switch p.Action {
	case "ask":

		peers = append(peers, e.GetRemoteAddr())
		p.Action = "reply"
		p.Peers = peers
		data, err := json.Marshal(p)
		if err != nil {
			return err
		}

		err = network.Sender(bytes.NewReader(data), network.DataType_Text, network.LADDR, e.GetRemoteAddr())
		return err

	case "reply":
		peers = append(peers, p.Peers...)
		return
	}

	return
}
