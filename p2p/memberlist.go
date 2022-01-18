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
	errlog     = []string{}
	localName  string
)

type ActionsType string

const (
	ActionsType_PUT ActionsType = "put"
	ActionsType_DEL ActionsType = "del"
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

	fmt.Println("msg:", string(b))

	switch b[0] {
	case 'd': // data
		var updates []*update
		if err := json.Unmarshal(b[1:], &updates); err != nil {
			return
		}
		for _, u := range updates {
			lc.Witness(u.Lt) ////然后比较收到消息的lamport时间和本地lamport时间，取大值（如果消息的大，那么+1）
			fmt.Println("cur lamport time:", lc.Time())
			err := writeLocaldb(u.Action, u.Data)
			if err != nil {
				return
			}
		}

	default:
		fmt.Println("other msg:", string(b))
	}
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return broadcasts.GetBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	m, err := readLocaldb("", 0)
	if err != nil {
		fmt.Println("get state error:", err)
	}
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
	m := [][]string{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return
	}

	//将合并的lamport时间读取到本地lamport时间
	for _, v := range m {
		if v[0] == "__lamporttime__" {
			otherlt, _ := strconv.ParseUint(v[1], 10, 64)
			lc.Witness(LamportTime(otherlt))
		}
	}

	err := writeLocaldb(ActionsType_PUT, m)
	if err != nil {
		fmt.Println("MergeRemoteState", err)
	}

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

func Start(clusterName string, port int, members []string) error {

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
	localName = c.Name

	mdnsInfo, err = stratMDNS(os.Stdout, c.Name, clusterName, nil, node.Addr, node.Port)
	if err != nil {
		return err
	}

	return nil
}

//处理发送消息
func SendMessage(action ActionsType, data [][]string, to ...memberlist.Address) (err error) {
	lt := lc.Increment() //发生发送事件，lamport时间+1 并附带到发送的消息中
	err = writeLocaldb(action, data)
	if err != nil {
		return
	}

	b, err := json.Marshal([]*update{
		{
			Action: action,
			Data:   data,
			Lt:     lt,
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

func writeLocaldb(action ActionsType, data [][]string) (err error) {

	mtx.Lock()
	defer mtx.Unlock()

	saveLamportTime() //保存lamport时间到数据库

	for _, v := range data {

		if len(v) != 2 {
			continue
		}

		switch action {
		case ActionsType_PUT:
			err = db.Put([]byte(v[0]), []byte(v[1]), nil)
		case ActionsType_DEL:
			err = db.Delete([]byte(v[0]), nil)
		}

		if err != nil {
			errlog = append(errlog, err.Error())
			break
		}

	}

	return
}

func readLocaldb(prefix string, limit uint) (m [][]string, err error) {

	iter := db.NewIterator(nil, nil)
	var i uint = 0
	for iter.Next() {
		i++
		if bytes.HasPrefix(iter.Key(), []byte(prefix)) || prefix == "" {
			m = append(m, []string{string(iter.Key()), string(iter.Value())})
		}
		if limit > 0 && i > limit {
			break
		}
	}

	iter.Release()
	err = iter.Error()
	return
}

//从数据库中提取lamporttime
func initLamportTime() {
	ltdata, err := db.Get([]byte("__lamporttime__"), nil)
	if err != nil {
		errlog = append(errlog, err.Error())
		return
	}
	lc.counter, err = strconv.ParseUint(string(ltdata), 10, 64)
	if err != nil {
		errlog = append(errlog, err.Error())
		return
	}
}

//保存lamporttime到数据库
func saveLamportTime() {
	err := db.Put([]byte("__lamporttime__"), []byte(fmt.Sprintf("%d", lc.Time())), nil)
	if err != nil {
		errlog = append(errlog, err.Error())
	}
}
