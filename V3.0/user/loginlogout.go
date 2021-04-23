//Package user contains all the registered user for the application and login logout feature.
package user

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"goInAction2Assignment/log"

	validator "github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

//User struct is a type that stored all registered user information.
type User struct {
	FirstName string
	LastName  string
	UserName  string
	Password  []byte
}

var (
	MapUsers    = map[string]User{}   //mapUsers uses username as a key and map with all its related information in user struct.
	MapSessions = map[string]string{} //mapSessions map registered user to a particular cookie value.
	tpl         *template.Template

	//Unique policy creation for the life of the program.
	Policy = bluemonday.UGCPolicy()
	// use a single instance of Validate
	validate *validator.Validate
)

func init() {

	//read all the gohtml file in the templates folder
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
	validate = validator.New()

}

//GenerateUser initiate preliminary user list and brcrypt their password in a secure manner.
func GenerateUser() {

	//registeredUsers was stored in json file encrypted.
	//For demo purpose, below is a few pair of username and password.
	// username: admin, password: password
	// username: ongryan, password: ongryan
	// username: tanalison, password: tanalison
	// all username password is set the same as the username
	fileout, _ := ioutil.ReadFile("user/registeredUsers.json")
	err := json.Unmarshal([]byte(fileout), &MapUsers)
	if err != nil {
		log.Fatal.Fatalln("Error in unmarshal registered user json file -", err)
	}
}

//Signup function enabled all the unregistered user to signup before they are able to access registered user feature.
func Signup(res http.ResponseWriter, req *http.Request) {

	if AlreadyLoggedIn(req) {
		http.Redirect(res, req, "/menu", http.StatusSeeOther)
		return
	}
	var myUser User

	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		firstname := strings.Title(req.FormValue("firstname"))
		err1 := validate.Var(firstname, "required,min=3,max=30,alphanum")
		if err1 != nil {
			http.Error(res, "Invalid firstname, please try again.", http.StatusForbidden)
			log.Warning.Println("Attempt to signup with invalid first name.-", err1)
		}

		lastname := strings.Title(req.FormValue("lastname"))
		err2 := validate.Var(lastname, "required,min=3,max=30,alphanum")
		if err2 != nil {
			http.Error(res, "Invalid lastname, please try again.", http.StatusForbidden)
			log.Warning.Println("Attempt to signup with invalid last name.-", err2)
		}

		username := strings.ToLower(req.FormValue("username"))
		err3 := validate.Var(username, "required,min=3,max=30,alphanum")
		if err3 != nil {
			http.Error(res, "Invalid username, please try again.", http.StatusForbidden)
			log.Warning.Println("Attempt to signup with invalid username.-", err3)
		}

		password := req.FormValue("password")
		err4 := validate.Var(password, "required,min=6,max=20,alphanum")
		if err4 != nil {
			http.Error(res, "Invalid password", http.StatusForbidden)
			log.Warning.Println("Attempt to signup with invalid password.-", err4)
		}

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			return
		}

		firstname = Policy.Sanitize(firstname)
		lastname = Policy.Sanitize(lastname)
		username = Policy.Sanitize(username)
		password = Policy.Sanitize(password)

		if username != "" {
			// check if username exist/ taken
			if _, ok := MapUsers[username]; ok {
				http.Error(res, "Username already taken, please login or choose a different username", http.StatusForbidden)
				log.Warning.Println("Attempt to signup an registered user.")
				return
			}
			// create session
			id := uuid.NewV4()
			myCookie := &http.Cookie{
				Name:     "myCookie",
				Value:    id.String(),
				Expires:  time.Now().Add(30 * time.Minute),
				HttpOnly: true,
				Path:     "/",
				Domain:   "127.0.0.1",
				Secure:   true,
			}
			http.SetCookie(res, myCookie)
			MapSessions[myCookie.Value] = username

			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				log.Error.Println("Internal server error - password hashing ", err)
				return
			}

			myUser = User{firstname, lastname, username, bPassword}
			MapUsers[username] = myUser
			file, _ := json.MarshalIndent(MapUsers, "", " ")
			_ = ioutil.WriteFile("user/registeredUsers.json", file, 0644)

			log.Info.Println("New user signed up.")
			//fmt.Println(mapUsers)
			//fmt.Println(mapSessions)
		}
		// redirect to main index
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return

	}

	err := tpl.ExecuteTemplate(res, "signup.gohtml", myUser)
	if err != nil {
		log.Fatal.Fatalln(err)
	}
}

//Login function allow registered user to login to the system.
//This function will also check for the existence of multiple login. Concurrent/mulitple login is not allowed for this apps.
func Login(res http.ResponseWriter, req *http.Request) {

	if AlreadyLoggedIn(req) {
		http.Redirect(res, req, "/menu", http.StatusSeeOther)
		return
	}

	// process form submission
	if req.Method == http.MethodPost {

		username := strings.ToLower(req.FormValue("username"))
		err1 := validate.Var(username, "required,min=3,max=30,alphanum")
		if err1 != nil {
			http.Error(res, "Username and/or password do not match", http.StatusForbidden)
			log.Warning.Println("Attempt to login with invalid username.-", err1)
			return
		}

		password := req.FormValue("password")
		err2 := validate.Var(password, "required,min=6,max=20,alphanum")
		if err2 != nil {
			http.Error(res, "Username and/or password do not match", http.StatusForbidden)
			log.Warning.Println("Attempt to login with invalid password.-", err2)
			return
		}

		username = Policy.Sanitize(username)
		password = Policy.Sanitize(password)

		// check if user exist with username
		myUser, ok := MapUsers[username]
		if !ok {
			http.Error(res, "Username and/or password do not match", http.StatusUnauthorized)
			log.Warning.Println("Failed login attempt.")
			return
		}
		// Matching of password entered
		err := bcrypt.CompareHashAndPassword(myUser.Password, []byte(password))
		if err != nil {
			http.Error(res, "Username and/or password do not match", http.StatusForbidden)
			log.Warning.Println("Failed login attempt.")
			return
		}
		//check if there is any existing current login, if yes, log out current browser and delete session
		if multiLogin(username) {
			http.Error(res, "Multiple login not allowed, please log out from other device. If you do not log in the account, please contact the adminstrator", http.StatusUnauthorized)
			log.Warning.Println("Multi login attempt is detected")
			return
		}

		// create session
		id := uuid.NewV4()
		myCookie := &http.Cookie{
			Name:     "myCookie",
			Value:    id.String(),
			Expires:  time.Now().Add(30 * time.Minute),
			HttpOnly: true,
			Path:     "/",
			Domain:   "127.0.0.1",
			Secure:   true,
		}
		http.SetCookie(res, myCookie)
		MapSessions[myCookie.Value] = username
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	err := tpl.ExecuteTemplate(res, "login.gohtml", nil)
	if err != nil {
		log.Fatal.Fatalln(err)
	}
}

//multiLogin is to check through all the sessions to see if particular user login twice
func multiLogin(username string) bool {

	for _, v := range MapSessions {
		if v == username {
			return true
		}
	}

	return false
}

//Logout function is able to access by user in all the pages. Once registered user is log out, the current session cookies will be cleared.
func Logout(res http.ResponseWriter, req *http.Request) {

	if !AlreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	myCookie, _ := req.Cookie("myCookie")
	// delete the session
	delete(MapSessions, myCookie.Value)
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
		log.Fatal.Fatalln(err)
	}
}

//GetUser function can helped to check if a particular session is map to any registered user.
//If an cookie value is able to map to registered user, it will return it. Else, the return would be empty.
func GetUser(res http.ResponseWriter, req *http.Request) User {
	// get current session cookie
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		id := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:     "myCookie",
			Value:    id.String(),
			HttpOnly: true,
			Expires:  time.Now().Add(30 * time.Minute),
			Path:     "/",
			Domain:   "127.0.0.1",
			Secure:   true,
		}
		http.SetCookie(res, myCookie)
	}

	// if the user exists already, get user
	var myUser User
	if username, ok := MapSessions[myCookie.Value]; ok {
		myUser = MapUsers[username]
	}

	return myUser
}

//AlreadyLoggedIn function enabled quick check of if a particular session is checked in.
func AlreadyLoggedIn(req *http.Request) bool {
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		return false
	}
	username := MapSessions[myCookie.Value]
	_, ok := MapUsers[username]
	return ok
}
