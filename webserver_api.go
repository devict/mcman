package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

/* JSON Functions */
func getJSON(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	w.Header().Set("Content-Type", "application/json")
	return mid(w, r)
}

func getAuthJSON(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	session, _ := sessionStore.Get(r, "mcman_session")

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

func getOnlineUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var err error
	var onlineUsers []MCUser
	if onlineUsers, err = c.model.getOnlineMCUsers(); err != nil {
		fmt.Fprintf(w, "{\"status\":\"error\",\"message\":\"Error loading online users\"}")
	}
	if err = json.NewEncoder(w).Encode(onlineUsers); err != nil {
		log.Println(err)
	}
}

func getOps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(c.Ops); err != nil {
		log.Println(err)
	}
}

func getWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(c.Whitelist); err != nil {
		log.Println(err)
	}
}
