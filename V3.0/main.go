//goInAction2Assignment is an application of a function room booking system for coworking space in Singapore.
//		Function of the application including:
//		1. Login/logout
//		2. Browse venue
//		3. Search venue
//		4. Book venue
//		5. Browse booking
//		6. Delete booking
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"goInAction2Assignment/booking"
	"goInAction2Assignment/log"
	"goInAction2Assignment/user"
	"goInAction2Assignment/venue"

	"github.com/gorilla/mux"
)

var tpl *template.Template

func init() {

	logChecksum, err := ioutil.ReadFile("log/checksumcommonlog.txt")
	if err != nil {
		fmt.Println(err)
	}

	str := string(logChecksum)

	if b, err := ComputeSHA256("log/commonlog.txt"); err != nil {
		fmt.Printf("Err: %v", err)
	} else {
		hash := hex.EncodeToString(b)
		if str == hash {
			log.Info.Println("Log integrity of common log is OK.")
		} else {
			log.Error.Println("File Tampering of common log detected.")
		}
	}

	logChecksum1, err := ioutil.ReadFile("log/checksumerrorslog.txt")
	if err != nil {
		fmt.Println(err)
	}

	str1 := string(logChecksum1)

	if b, err := ComputeSHA256("log/errorslog.txt"); err != nil {
		fmt.Printf("Err: %v", err)
	} else {
		hash := hex.EncodeToString(b)
		if str1 == hash {
			log.Info.Println("Log integrity of error log is OK.")
		} else {
			log.Error.Println("File Tampering of error log detected.")
		}
	}
}

func init() {

	venue.GenerateSlots()
	user.GenerateUser()
	err := booking.GenerateBookings()
	if err != nil {
		log.Error.Println(err)
	}
	//pre load some information during the start of application for better illustration
}

func init() {

	//read all the gohtml file in the templates folder
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

}

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", Index).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/menu", menu).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/browseVenue", venue.BrowseVenue).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/searchVenue", venue.SearchVenue).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/bookVenue", booking.BookVenue).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/removeBooking", booking.MainBookingList.RemoveBooking).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/browseBooking", booking.MainBookingList.BrowseBooking).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/signup", user.Signup).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/login", user.Login).Methods("GET", "POST").Schemes("https")
	r.HandleFunc("/logout", user.Logout).Methods("GET", "POST").Schemes("https")
	r.Handle("/favicon.ico", http.NotFoundHandler())
	
	//connect to the port according to the assignment requirement
	err := http.ListenAndServeTLS(":5221", "cert/cert.pem", "cert/key.pem", r)
	if err != nil {
		log.Fatal.Fatalln("ListenAndServe: ", err)
	}

}

//Index is welcome page of the application, visitor can choose the mode of browsing.
func Index(res http.ResponseWriter, req *http.Request) {

	myUser := user.GetUser(res, req)
	err := tpl.ExecuteTemplate(res, "index.gohtml", myUser)
	if err != nil {
		log.Fatal.Fatalln(err)
	}

}

//Based of their mode of browsing, different customise menu will be presented.
func menu(res http.ResponseWriter, req *http.Request) {

	myUser := user.GetUser(res, req)
	err := tpl.ExecuteTemplate(res, "menu.gohtml", myUser)
	if err != nil {
		log.Fatal.Fatalln(err)
	}
}

//ComputeSHA256 compute the checksum of the logfile to ensure the logfile has not been tempered when it is opened in the application.
func ComputeSHA256(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}
