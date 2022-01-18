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

	"github.com/hashicorp/memberlist"
)

func putHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("val")

	err := SendMessage(ActionsType_PUT, [][]string{{key, val}})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("put success"))
}

func directlineHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	val := r.Form.Get("val")

	k := fmt.Sprintf("LINE_L:%011d_N:%s", lc.Time(), localName)
	fmt.Println(k, val)
	err := SendMessage(ActionsType_PUT, [][]string{{k, val}})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("put line success"))
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	err := SendMessage(ActionsType_DEL, [][]string{{key, ""}})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte("del success"))
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	val, err := db.Get([]byte(key), nil)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(val))
}

func kv(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	prefix := r.Form.Get("prefix")
	seek := r.Form.Get("seek")

	fmt.Println("request seek:", seek, prefix)

	m, err := readLocaldb(prefix, seek, 0)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	b, _ := json.Marshal(m)
	w.Write(b)
}

func sendtoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	toIP := r.Form.Get("to_ip")
	toName := r.Form.Get("to_name")
	key := r.Form.Get("key")
	val := r.Form.Get("val")

	err := SendMessage(ActionsType_PUT, [][]string{{key, val}}, memberlist.Address{Name: toName, Addr: toIP})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("send success"))
}

func joinHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	member := r.Form.Get("member")

	i, err := memberList.Join([]string{member})

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(fmt.Sprintf("join success: %d", i)))
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
	members := []string{}
	localName = r.Form.Get("local_name")
	clusterName := r.Form.Get("cluster_name")
	portStr := r.Form.Get("port")
	member := r.Form.Get("member")

	if localName == "" {
		localName, _ = os.Hostname()
	}
	if clusterName == "" {
		clusterName = "mycluster"
	}
	if member != "" {
		members = append(members, member)
	}

	port, _ := strconv.Atoi(portStr)

	if err := Start(clusterName, port, members); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("start success"))
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

	w.Write([]byte("stop success"))
}

func errorlog(w http.ResponseWriter, r *http.Request) {

	errlogJson, err := json.Marshal(errlog)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Write([]byte(errlogJson))
}

func dashboard(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./web/dashboard.html")
	if err != nil {
		panic(err)
	}

	t.Execute(w, nil)
}

func chat(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./web/chat.html")
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
