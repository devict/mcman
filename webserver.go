package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

var output_channel chan string

type WebUser struct {
	Username string
	Password string
}

var session_store = sessions.NewCookieStore([]byte("super_secret_secret :D"))

func StartServer(ch chan string) {
	output_channel = ch
	_, err := os.Stat("mapcrafter/index.html")
	if err == nil {
		// Looks like mapcrafter is present
		output_channel <- "* Mapcrafter Directory is Present, routing to /\n"
		fs := http.FileServer(http.Dir("mapcrafter"))
		http.Handle("/", fs)
	}

	// Expose the public directory for CSS / JS assets. We could just use
	// http.Handle on "/" and not do the StripPrefix but then the tmpls/
	// directory would also be accessible.
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(Dir(true, "/public"))))

	http.HandleFunc("/admin/", serveMcMan)
	output_channel <- "* Admin site running at /admin/\n"
	http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
}

func serveMcMan(w http.ResponseWriter, r *http.Request) {
	output := ""
	output_channel <- fmt.Sprint("HTTP Request: (", r.Method, ") ", r.URL, "\n")

	the_path := r.URL.Path
	the_path = strings.TrimPrefix(strings.ToLower(the_path), "/admin")

	if strings.HasPrefix(the_path, "/login") || the_path == "/" {
		output = getHTML(loginScreen, w, r)
	} else if strings.HasPrefix(the_path, "/dologin") {
		output = getHTML(doLogin, w, r)
	} else if strings.HasPrefix(the_path, "/api/") {
		output = getAuthJSON(serveAPI, w, r)
	}
	fmt.Fprintf(w, output)
}

func serveAPI(w http.ResponseWriter, r *http.Request) string {
	_, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	//body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// What are we doing with this request?
	the_path := strings.TrimPrefix(strings.ToLower(r.URL.Path), "/admin")
	output_string := ""
	if strings.HasPrefix(the_path, "/api") {
		the_path = strings.TrimPrefix(the_path, "/api")
		if strings.HasPrefix(the_path, "/v1") {
			the_path = strings.TrimPrefix(the_path, "/v1")
			if strings.HasPrefix(the_path, "/whitelist") {
				output_string = handleWhitelist(r)
			} else if strings.HasPrefix(the_path, "/ops") {
				output_string = handleOps(r)
			} else if strings.HasPrefix(the_path, "/stop") {
				DoStopServer()
			} else if strings.HasPrefix(the_path, "/init") {
				v := WebUser{"br0xen", "asdf"}
				c.model.updateWebUser(&v)
			}
		}
	}
	return output_string
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

func handleOps(r *http.Request) string {
	if r.Method == "GET" {
		return getOps()
	} else if r.Method == "POST" {
		// Add posted user to Ops
	}
	return ""
}

func handleWhitelist(r *http.Request) string {
	if r.Method == "GET" {
		return getWhitelist()
	} else if r.Method == "POST" {
		// Add posted user to whitelist
	}
	return ""
}

/* HTML Functions */
func getAuthHTML(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	session, _ := session_store.Get(r, "mcman_session")

	// Vaildate that the session is authenticated
	if val, ok := session.Values["is_logged_in"].(string); ok {
		switch val {
		case "":
			http.Redirect(w, r, "/admin/login", http.StatusFound)
		default:
			return mid(w, r)
		}
	} else {
		http.Redirect(w, r, "/admin/login", http.StatusFound)
	}
	return ""
}

func getHTML(mid func(http.ResponseWriter, *http.Request) string,
	w http.ResponseWriter, r *http.Request) string {
	return fmt.Sprintf("%s%s%s", htmlHeader("Minecraft Manager"), mid(w, r), htmlFooter())
}

func loginScreen(w http.ResponseWriter, r *http.Request) string {
	return `
	<form class="pure-form" action="/admin/doLogin" method="POST">
		<fieldset>
			<input type="text" name="username" placeholder="username">
			<input type="password" name="password" placeholder="password">
			<button type="submit" class="pure-button pure-button-primary button-success">sign in</button>
		</fieldset>
	</form>
`
}

func doLogin(w http.ResponseWriter, r *http.Request) string {
	ret := "Do Login<br />"
	login_user := r.FormValue("username")
	login_pass := r.FormValue("password")
	lu := c.model.getWebUser(login_user)
	// Set the Cookie

	if login_pass == lu.Password {
		session, _ := session_store.Get(r, "mcman_session")
		session.Values["is_logged_in"] = login_user
		session.Save(r, w)

		ret = ret + "Logged In!"
	}

	return ret
}

func htmlHeader(title string) string {
	head := `
<!DOCTYPE html>
<html>
	<head>
		<title>`
	head += title
	head += `
		</title>
		<link rel="stylesheet" href="/public/css/pure.css">
		<link rel="stylesheet" href="/public/css/mcman.css">
<!--[if lte IE 8]>
    <link rel="stylesheet" href="http://yui.yahooapis.com/pure/0.6.0/grids-responsive-old-ie-min.css">
<![endif]-->
<!--[if gt IE 8]><!-->
    <link rel="stylesheet" href="http://yui.yahooapis.com/pure/0.6.0/grids-responsive-min.css">
<!--<![endif]-->
	</head>
	<body>
	<div class="mcman_wrapper pure-g" id="menu">
		<div class="pure-u-1 pure-u-md-1-3">
			<div class="pure-menu">
				<a href="#" class="pure-menu-heading custom-brand">Buller Mineworks</a>
				<a href="#" class="mcman-toggle" id="toggle"><s class="bar"></s><s class="bar"></s></a>
			</div>
		</div>
		<div class="pure-u-1 pure-u-md-1-3">
			<div class="pure-menu pure-menu-horizontal mcman-can-transform">
				<ul class="pure-menu-list">
					<li class="pure-menu-item"><a href="#" class="pure-menu-link">Home</a></li>
					<li class="pure-menu-item"><a id="stop_link" href="#" class="pure-menu-link">Stop</a></li>
					<li class="pure-menu-item"><a href="#" class="pure-menu-link">About</a></li>
				</ul>
			</div>
		</div>
		<div class="pure-u-1 pure-u-md-1-3">
			<div class="pure-menu pure-menu-horizontal mcman-menu-3 mcman-can-transform">
				<ul class="pure-menu-list">
					<li class="pure-menu-item"><a href="/" class="pure-menu-link">Map</a></li>
				</ul>
			</div>
		</div>
	</div>
`
	return head
}

func htmlFooter() string {
	return `
	<script src="/public/js/B.js"></script>
	<script src="/public/js/mcman.js"></script>
	</body>
</html>
`
}

/* Data Functions */
func getOps() string {
	ret := "["
	num_users := 0
	for _, op_user := range GetConfig().Ops {
		if num_users > 0 {
			ret += ","
		}
		ret += fmt.Sprint("\"", op_user, "\"")
	}
	ret += "]"
	return ret
}

func getWhitelist() string {
	ret := "["
	num_users := 0
	for _, wl_user := range GetConfig().Whitelist {
		if num_users > 0 {
			ret += ","
		}
		ret += fmt.Sprint("\"", wl_user, "\"")
	}
	ret += "]"
	return ret
}
