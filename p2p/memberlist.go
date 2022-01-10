package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/pborman/uuid"
)

var (
	mtx        sync.RWMutex
	items      = map[string]string{} //存储，此处需要修改为持久性存储
	broadcasts *memberlist.TransmitLimitedQueue
	memberList *memberlist.Memberlist
	mdnsInfo   *agentMDNS
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct{}

type update struct {
	Action string // add, del
	Data   map[string]string
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	if len(b) == 0 {
		return
	}

	switch b[0] {
	case 'd': // data
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		mtx.Lock()
		for _, u := range updates {
			for k, v := range u.Data {
				switch u.Action {
				case "add":
					items[k] = v
				case "del":
					delete(items, k)
				}
			}
		}
		mtx.Unlock()
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	mtx.RLock()
	m := items
	mtx.RUnlock()
	b, _ := json.Marshal(m)
	return b
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
	if len(buf) == 0 {
		return
	}
	if !join {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}
	mtx.Lock()
	for k, v := range m {
		items[k] = v
	}
	mtx.Unlock()
}

type eventDelegate struct{}

func (ed *eventDelegate) NotifyJoin(node *memberlist.Node) {
	fmt.Println("A node has joined: " + node.String())
}

func (ed *eventDelegate) NotifyLeave(node *memberlist.Node) {
	fmt.Println("A node has left: " + node.String())
}

func (ed *eventDelegate) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("A node was updated: " + node.String())
}

func Start(members []string) error {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLocalConfig()
	c.Events = &eventDelegate{}
	c.Delegate = &delegate{}
	c.BindPort = 0
	c.Name = hostname + "-" + uuid.NewUUID().String()
	var err error
	memberList, err = memberlist.Create(c)
	if err != nil {
		return err
	}
	if len(members) > 0 {
		_, err := memberList.Join(members)
		if err != nil {
			return err
		}
	}
	broadcasts = &memberlist.TransmitLimitedQueue{
		NumNodes: func() int {
			return memberList.NumMembers()
		},
		RetransmitMult: 3,
	}
	node := memberList.LocalNode()
	fmt.Printf("Local member %s:%d\n", node.Addr, node.Port)

	mdnsInfo, err = stratMDNS(os.Stdout, c.Name, "metanet", nil, node.Addr, node.Port)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}
