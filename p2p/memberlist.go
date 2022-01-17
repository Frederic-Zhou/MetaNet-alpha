package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

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
	errlog     = []string{}
)

type ActionsType string

const (
	ActionsType_PUT  ActionsType = "put"
	ActionsType_DEL  ActionsType = "del"
	ActionsType_LINE ActionsType = "line"
)

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

type delegate struct{}

type update struct {
	Action ActionsType // put, del
	Data   [][]string
	Lt     LamportTime
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

//处理收到消息
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
				var err error
				switch u.Action {
				case ActionsType_PUT:
					// data := fmt.Sprintf("%s,%d", v, lc.Time())
					err = db.Put([]byte(v[0]), []byte(v[1]), nil)
				case ActionsType_LINE:
					dataByte, _ := json.Marshal(v)
					err = db.Put([]byte(fmt.Sprintf("line_t%d_l%d_i%s", time.Now().Unix(), lc.Time(), k)), []byte(dataByte), nil)
				case ActionsType_DEL:
					err = db.Delete([]byte(v[1]), nil)
				}

				if err != nil {
					errlog = append(errlog, err.Error())
				}
			}
		}

		mtx.Unlock()
	default:
		fmt.Println("other msg:", string(b))
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
		err := db.Put([]byte(k), []byte(v), nil)
		if err != nil {
			errlog = append(errlog, err.Error())
		}

		// i := bytes.LastIndex([]byte(v), []byte(","))
		// t, _ := strconv.ParseUint(string(v[i+1:]), 10, 64)
		// lc.Witness(LamportTime(t) - 1)
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

//处理发送消息
func SendMessage(action ActionsType, data [][]string, to ...memberlist.Address) (err error) {

	for k, v := range data {
		switch action {
		case ActionsType_PUT:
			// data := fmt.Sprintf("%s,%d", v, lc.Time())
			err = db.Put([]byte(v[0]), []byte(v[1]), nil)
		case ActionsType_LINE:
			dataByte, _ := json.Marshal(v)
			err = db.Put([]byte(fmt.Sprintf("line_t%d_l%d_i%s", time.Now().Unix(), lc.Time(), k)), []byte(dataByte), nil)
		case ActionsType_DEL:
			err = db.Delete([]byte(v[1]), nil)
		}

		if err != nil {
			errlog = append(errlog, err.Error())
		}

	}

	b, err := json.Marshal([]*update{
		{
			Action: action,
			Data:   data,
			Lt:     lc.Increment(),
		},
	})

	if err != nil {
		return
	}

	//有 to 单发
	if len(to) > 0 {

		for _, toAddr := range to {
			err = memberList.SendToAddress(toAddr, append([]byte("d"), b...))
			if err != nil {
				return
			}
		}

	} else { //无 to 广播
		broadcasts.QueueBroadcast(&broadcast{
			msg:    append([]byte("d"), b...),
			notify: nil,
		})
	}

	return
}
