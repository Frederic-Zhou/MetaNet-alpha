package handler

import (
	"encoding/json"

	"github.com/Frederic-Zhou/MetaNet-alpha/NAT2P/network"
	"github.com/sirupsen/logrus"
)

type Msg struct {
	Content string `json:"content"`
}

func (m Msg) Do(e *network.Event) (err error) {

	err = json.Unmarshal(e.GetBody(), &m)
	if err != nil {
		return
	}

	if m.Content != "" {
		logrus.Infof("[%s]\n", m.Content)
	}

	return
}
