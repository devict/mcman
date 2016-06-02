package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type WebUser struct {
	Username string
	Password string
}

var session_store = sessions.NewCookieStore([]byte("super_secret_secret :D"))

var v view

func StartServer(dev bool) {
	v = newView(dev)

	r := mux.NewRouter()

	// Expose the public directory for CSS / JS assets. We could just use
	// http.Handle on "/" and not do the StripPrefix but then the tmpls/
	// directory would also be accessible. The "Dir" function is generated by
	// the esc tool from "go generate"
	r.PathPrefix("/public").Handler(
		http.StripPrefix("/public", http.FileServer(Dir(dev, "/public"))),
	)

	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/login", login).Methods("GET")
	r.HandleFunc("/dologin", doLogin).Methods("GET")

	a := r.PathPrefix("/api/v1").Subrouter()

	a.HandleFunc("/whitelist", getWhitelist).Methods("GET")
	a.HandleFunc("/ops", getOps).Methods("GET")
	a.HandleFunc("/stop", postStop).Methods("POST")

	output <- "* Site running at localhost:8080\n"
	http.ListenAndServe(":8080", context.ClearHandler(r))
}

func index(w http.ResponseWriter, r *http.Request) {
	t := v.Tmpl("index")
	fmt.Println(t.Execute(w, nil))
}

func login(w http.ResponseWriter, r *http.Request) {
	t := v.Tmpl("login")
	fmt.Println(t.Execute(w, nil))
}

func doLogin(w http.ResponseWriter, r *http.Request) {
	login_user := r.FormValue("username")
	login_pass := r.FormValue("password")

	lu := c.model.getWebUser(login_user)

	if login_pass == lu.Password {
		session, _ := session_store.Get(r, "mcman_session")
		session.Values["is_logged_in"] = login_user
		session.Save(r, w)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func serveAPI(w http.ResponseWriter, r *http.Request) string {
	return ""

	//if strings.HasPrefix(the_path, "/init") {
	//v := WebUser{"br0xen", "asdf"}
	//c.model.updateWebUser(&v)
	//}
}

/* JSON Functions */
func getJSON(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	w.Header().Set("Content-Type", "application/json")
	return mid(w, r)
}

func getAuthJSON(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	session, _ := session_store.Get(r, "mcman_session")

	// Vaildate that the session is authenticated
	if val, ok := session.Values["is_logged_in"].(string); ok {
		switch val {
		case "":
			w.WriteHeader(403)
			w.Header().Set("Content-Type", "application/json")
			return "{\"status\":\"error\"}"
		default:
			w.Header().Set("Content-Type", "application/json")
			return mid(w, r)
		}
	} else {
		w.WriteHeader(403)
		w.Header().Set("Content-Type", "application/json")
		return "{\"status\":\"error\"}"
	}
	return ""
}

func getOps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(GetConfig().Ops); err != nil {
		log.Println(err)
	}
}

func getWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(GetConfig().Whitelist); err != nil {
		log.Println(err)
	}
}

func postStop(w http.ResponseWriter, r *http.Request) {
	DoStopServer()
}
