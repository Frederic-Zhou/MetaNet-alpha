package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

func putHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("val")
	mtx.Lock()
	db.Put([]byte(key), []byte(val), nil)
	mtx.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: "put",
			Data: map[string]string{
				key: val,
			},
		},
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})

	w.Write([]byte("put success"))
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	mtx.Lock()
	db.Delete([]byte(key), nil)
	mtx.Unlock()

	b, err := json.Marshal([]*update{{
		Action: "del",
		Data: map[string]string{
			key: "",
		},
	}})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	broadcasts.QueueBroadcast(&broadcast{
		msg:    append([]byte("d"), b...),
		notify: nil,
	})

	w.Write([]byte("del success"))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	mtx.RLock()
	// val := items[key]
	val, err := db.Get([]byte(key), nil)
	mtx.RUnlock()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(val))
}

func kv(w http.ResponseWriter, r *http.Request) {
	mtx.RLock()
	m := map[string]string{}
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		m[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()
	mtx.RUnlock()

	if err := iter.Error(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	b, _ := json.Marshal(m)
	w.Write(b)
}

func joinHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	member := r.Form.Get("member")

	i, err := memberList.Join([]string{member})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(fmt.Sprintf("%d", i)))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	info, err := json.Marshal(map[string]interface{}{
		"health_score": memberList.GetHealthScore(),
		"members":      memberList.Members(),
		"mdns":         *mdnsInfo,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write(info)

}

func start(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	localName := r.Form.Get("local_name")
	clusterName := r.Form.Get("cluster_name")
	portStr := r.Form.Get("port")
	member := r.Form.Get("member")

	if localName == "" {
		localName, _ = os.Hostname()
	}
	if clusterName == "" {
		clusterName = "mycluster"
	}

	port, _ := strconv.Atoi(portStr)

	if err := Start(localName, clusterName, port, []string{member}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("success"))
}

func stop(w http.ResponseWriter, r *http.Request) {

	err := mdnsInfo.Server.Shutdown()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	err = memberList.Shutdown()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("success"))
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./web/dashboard.html")
	if err != nil {
		panic(err)
	}

	t.Execute(w, nil)
}
