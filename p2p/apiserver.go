package main

import (
	"crypto/sha256"
	"crypto/subtle"
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

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the username and password from the request
		// Authorization header. If no Authentication header is present
		// or the header value is invalid, then the 'ok' return value
		// will be false.
		username, password, ok := r.BasicAuth()
		if ok {
			// Calculate SHA-256 hashes for the provided and expected
			// usernames and passwords.
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte("zeta"))
			expectedPasswordHash := sha256.Sum256([]byte("chow"))

			// Use the subtle.ConstantTimeCompare() function to check if
			// the provided username and password hashes equal the
			// expected username and password hashes. ConstantTimeCompare
			// will return 1 if the values are equal, or 0 otherwise.
			// Importantly, we should to do the work to evaluate both the
			// username and password before checking the return values to
			// avoid leaking information.
			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			// If the username and password are correct, then call
			// the next handler in the chain. Make sure to return
			// afterwards, so that none of the code below is run.
			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		// If the Authentication header is not present, is invalid, or the
		// username or password is wrong, then set a WWW-Authenticate
		// header to inform the client that we expect them to use basic
		// authentication and send a 401 Unauthorized response.
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
