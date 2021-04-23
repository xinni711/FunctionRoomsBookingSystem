package main

import (
	"fmt"
	"log"
	"net/http"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

//generateUser initiate preliminary user list and brcrypt their password
func generateUser() {

	mPass, _ := bcrypt.GenerateFromPassword([]byte("leeken"), bcrypt.MinCost)
	mapUsers["lee.ken"] = user{"Ken", "Lee", "lee.ken", mPass}
	tPass, _ := bcrypt.GenerateFromPassword([]byte("tanalison"), bcrypt.MinCost)
	mapUsers["tan.alison"] = user{"Alison", "Tan", "tan.alison", tPass}
	lPass, _ := bcrypt.GenerateFromPassword([]byte("lim.christina"), bcrypt.MinCost)
	mapUsers["lim.christina"] = user{"Christina", "Lim", "lim.christina", lPass}
	oPass, _ := bcrypt.GenerateFromPassword([]byte("ongryan"), bcrypt.MinCost)
	mapUsers["ong.ryan"] = user{"Ryan", "Ong", "ong.ryan", oPass}

}

func signup(res http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/menu", http.StatusSeeOther)
		return
	}
	var myUser user
	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		firstname := req.FormValue("firstname")
		lastname := req.FormValue("lastname")
		username := req.FormValue("username")
		password := req.FormValue("password")

		if username != "" {
			// check if username exist/ taken
			if _, ok := mapUsers[username]; ok {
				http.Error(res, "Username already taken, please login or choose a different username", http.StatusForbidden)
				return
			}
			// create session
			id := uuid.NewV4()
			myCookie := &http.Cookie{
				Name:  "myCookie",
				Value: id.String(),
			}
			http.SetCookie(res, myCookie)
			mapSessions[myCookie.Value] = username

			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				return
			}

			myUser = user{firstname, lastname, username, bPassword}
			mapUsers[username] = myUser
			fmt.Println(mapUsers)
			fmt.Println(mapSessions)
		}
		// redirect to main index
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return

	}

	err := tpl.ExecuteTemplate(res, "signup.gohtml", myUser)
	if err != nil {
		log.Fatalln(err)
	}
}

func login(res http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/menu", http.StatusSeeOther)
		return
	}

	// process form submission
	if req.Method == http.MethodPost {
		username := req.FormValue("username")
		password := req.FormValue("password")
		// check if user exist with username
		myUser, ok := mapUsers[username]
		if !ok {
			http.Error(res, "Username and/or password do not match", http.StatusUnauthorized)
			return
		}
		// Matching of password entered
		err := bcrypt.CompareHashAndPassword(myUser.Password, []byte(password))
		if err != nil {
			http.Error(res, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// create session
		id := uuid.NewV4()
		myCookie := &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}
		http.SetCookie(res, myCookie)
		mapSessions[myCookie.Value] = username
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	err := tpl.ExecuteTemplate(res, "login.gohtml", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func logout(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	myCookie, _ := req.Cookie("myCookie")
	// delete the session
	delete(mapSessions, myCookie.Value)
	// remove the cookie
	myCookie = &http.Cookie{
		Name:   "myCookie",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(res, myCookie)

	http.Redirect(res, req, "/", http.StatusSeeOther)

	err := tpl.ExecuteTemplate(res, "login.gohtml", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func getUser(res http.ResponseWriter, req *http.Request) user {
	// get current session cookie
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		id := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}

	}
	http.SetCookie(res, myCookie)

	// if the user exists already, get user
	var myUser user
	if username, ok := mapSessions[myCookie.Value]; ok {
		myUser = mapUsers[username]
	}

	return myUser
}

func alreadyLoggedIn(req *http.Request) bool {
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		return false
	}
	username := mapSessions[myCookie.Value]
	_, ok := mapUsers[username]
	return ok
}
