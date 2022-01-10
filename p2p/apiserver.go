package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func addHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("val")
	mtx.Lock()
	items[key] = val
	mtx.Unlock()

	b, err := json.Marshal([]*update{
		{
			Action: "add",
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

	w.Write([]byte("add success"))
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	mtx.Lock()
	delete(items, key)
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
	val := items[key]
	mtx.RUnlock()
	w.Write([]byte(val))
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
		"kv":           items,
	})
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.Write(info)

}
