package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func initAdminRequest(w http.ResponseWriter, req *http.Request) *pageData {
	p := initPageData(w, req)
	p.Stylesheets = append(p.Stylesheets, "/assets/css/admin.css")
	p.Scripts = append(p.Scripts, "/assets/js/admin.js")

	return p
}

func handleAdmin(w http.ResponseWriter, req *http.Request) {
	var err error
	page := initAdminRequest(w, req)
	vars := mux.Vars(req)
	if !page.LoggedIn {
		page.SubTitle = "Admin Login"
		page.show("admin-login.html", w)
	} else {
		adminCategory := vars["category"]
		switch adminCategory {
		case "stop":
			handleAdminDoStop(w, req, page)
		case "users":
			handleAdminUsers(w, req, page)
		default:
			type indexData struct {
				OnlineUsers []MCUser
			}
			id := new(indexData)
			if id.OnlineUsers, err = c.model.getOnlineMCUsers(); err != nil {
				fmt.Printf("> Error loading users: " + err.Error())
			}
			page.TemplateData = id
			page.show("admin-main.html", w)
		}
	}
}

// handleAdminDoLogin
// Verify the provided credentials, set up a cookie (if requested)
// and redirect back to /admin
func handleAdminDoLogin(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	// Fetch the login credentials
	login_user := req.FormValue("username")
	login_pass := req.FormValue("password")

	err := c.model.checkWebUserCreds(login_user, login_pass)
	if err == nil {
		page.session.setFlashMessage("Logged in", "success")
		page.session.setStringValue("login", login_user)
	} else {
		page.session.setFlashMessage("Login failed", "error")
	}

	redirect("/admin", w, req)
}

func handleAdminDoLogout(w http.ResponseWriter, req *http.Request) {
	page := initAdminRequest(w, req)
	page.session.expireSession()
	page.session.setFlashMessage("Logged out", "success")

	redirect("/admin", w, req)
}

func handleAdminUsers(w http.ResponseWriter, req *http.Request, page *pageData) {
	type usrPage struct {
		Users []string
	}
	u := new(usrPage)
	u.Users = c.model.getAllWebUsers()
	page.TemplateData = u
	page.show("admin-users.html", w)
}

func handleAdminDoStop(w http.ResponseWriter, req *http.Request, page *pageData) {
	DoStopServer()
	page.show("stopped.html", w)
}
