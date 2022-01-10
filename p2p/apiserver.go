package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
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
	if err != nil {
		fmt.Println("get value error:", err)
	}
	mtx.RUnlock()
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

	if err := iter.Error(); err != nil {
		fmt.Println("get state error:", err)
	}
	mtx.RUnlock()
	b, _ := json.Marshal(m)
	w.Write(b)
}

func joinHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	member := r.Form.Get("member")

	i, err := memberList.Join([]string{member})

	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write([]byte(fmt.Sprintf("%d", i)))
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println(mdnsInfo)
	info, err := json.Marshal(map[string]interface{}{
		"health_score": memberList.GetHealthScore(),
		"members":      memberList.Members(),
		"mdns":         *mdnsInfo,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write(info)

}

func dashboard(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./dashboard.html")
	if err != nil {
		panic(err)
	}

	t.Execute(w, nil)
}
