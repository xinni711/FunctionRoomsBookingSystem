package main

import (
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

//user information are stored in struct form
type user struct {
	FirstName string
	LastName  string
	UserName  string
	Password  []byte
}

var mapUsers = map[string]user{}
var mapSessions = map[string]string{}
var tpl *template.Template

func init() {
	generateSlots()
	generateUser()
	generateBookings()
	//pre load some information during the start of application for better illustration
}

func init() {

	//read all the gohtml file in the templates folder
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
	bPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	mapUsers["admin"] = user{"Matthew", "Lee", "admin", bPassword}
	//initiate the admin
}

func main() {

	http.HandleFunc("/", index)
	http.HandleFunc("/menu", menu)
	http.HandleFunc("/browseVenue", browseVenue)
	http.HandleFunc("/searchVenue", searchVenue)
	http.HandleFunc("/bookVenue", bookVenue)
	http.HandleFunc("/removeBooking", MainBookingList.removeBooking)
	http.HandleFunc("/browseBooking", MainBookingList.browseBooking)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":5221", nil) //connect to the port according to the assignment requirement
}

//welcome page of the application, visitor can choose the mode of browsing
func index(res http.ResponseWriter, req *http.Request) {

	myUser := getUser(res, req)
	err := tpl.ExecuteTemplate(res, "index.gohtml", myUser)
	if err != nil {
		log.Fatalln(err)
	}

}

//based of their mode of browsing, different customise menu will be presented
func menu(res http.ResponseWriter, req *http.Request) {

	myUser := getUser(res, req)
	err := tpl.ExecuteTemplate(res, "menu.gohtml", myUser)
	if err != nil {
		log.Fatalln(err)
	}
}
