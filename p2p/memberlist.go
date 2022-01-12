package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/hashicorp/memberlist"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	mtx        sync.RWMutex
	broadcasts *memberlist.TransmitLimitedQueue
	memberList *memberlist.Memberlist
	mdnsInfo   *agentMDNS
	db         *leveldb.DB
	lc         = LamportClock{counter: 0}
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct{}

type update struct {
	Action      string // put, del
	Data        map[string]string
	Persistence bool
	Lt          LamportTime
}

func init() {

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
			lc.Witness(u.Lt)
			for k, v := range u.Data {
				switch u.Action {
				case "put":
					data := fmt.Sprintf("%s,%d", v, lc.Time())
					db.Put([]byte(k), []byte(data), nil)
				case "del":
					db.Delete([]byte(k), nil)
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
	m := map[string]string{}
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		m[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()

	if err := iter.Error(); err != nil {
		fmt.Println("get state error:", err)
	}
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
		db.Put([]byte(k), []byte(v), nil)

		i := bytes.LastIndex([]byte(v), []byte(","))
		t, _ := strconv.ParseUint(string(v[i+1:]), 10, 64)
		lc.Witness(LamportTime(t) - 1)
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

func Start(localName, clusterName string, port int, members []string) error {

	c := memberlist.DefaultLocalConfig()
	c.Events = &eventDelegate{}
	c.Delegate = &delegate{}
	c.BindPort = port
	c.Name = localName

	var err error
	memberList, err = memberlist.Create(c)
	if err != nil {
		fmt.Println("create err", err)
		return err
	}

	if len(members) > 0 {
		_, err := memberList.Join(members)
		if err != nil {
			fmt.Println("join err", err, members, len(members))
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

	mdnsInfo, err = stratMDNS(os.Stdout, c.Name, clusterName, nil, node.Addr, node.Port)
	if err != nil {
		return err
	}

	return nil
}
