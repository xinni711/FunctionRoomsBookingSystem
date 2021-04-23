package booking

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"goInAction2Assignment/log"
	"goInAction2Assignment/user"
	"goInAction2Assignment/venue"
)

//RemoveBooking function will perform remove booking feature based on booking ID. 
//Only admin can perform remove booking.He/she can select the booking ID based on the full booking list displayed below.
func (b *bookingList) RemoveBooking(res http.ResponseWriter, req *http.Request) {

	var bookingIDtoRemove int64

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			log.Panic.Println("Recovered from panic for removeBooking feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if !user.AlreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	myUser := user.GetUser(res, req)

	if myUser.UserName != "admin" {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {

		//obtain user input of booking ID to remove

		bookingIDtoremoveS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("bookingIDToDelete")))
		regexbookingIDtoremoveS := regexp.MustCompile(`^[0-9]{1,5}$`)
		if !regexbookingIDtoremoveS.MatchString(bookingIDtoremoveS) {
			http.Error(res, "Invalid bookingID selection, please try again.", http.StatusBadRequest)
			log.Warning.Println("Invalid bookingIDtoRemove selection.")
			return
		}
		bookingIDtoRemove, _ = strconv.ParseInt(bookingIDtoremoveS, 10, 64)

		if bookingIDtoRemove != 0 {

			currentBooking := b.head

			if b.head == nil {
				http.Error(res, "Empty booking list, nothing to be deleted", http.StatusBadRequest)
				log.Warning.Println("Empty booking list")
				return
			}

			if currentBooking.next == nil {
				if currentBooking.BookingID == bookingIDtoRemove {
					//remove booked slot, reset them back to available
					//Info.Println("Deleting booking ID", currentBooking.BookingID)
					err := venue.Room[currentBooking.BookingSlot[0]%100-1].Slot.RemoveBookedSlot(currentBooking.BookingSlot)
					if err != nil {
						http.Error(res, "Error removing booked slot", http.StatusBadRequest)
						log.Error.Println("Error removing booked slot")
					}
					b.head = nil
				}
			}

			for i := 1; i < b.size; i++ {
				prevUser := currentBooking
				currentBooking = currentBooking.next
				if currentBooking.BookingID == bookingIDtoRemove {
					//Info.Println("Deleting booking ID", currentBooking.BookingID)
					err := venue.Room[currentBooking.BookingSlot[0]%100-1].Slot.RemoveBookedSlot(currentBooking.BookingSlot)
					if err != nil {
						http.Error(res, "Error removing booked slot", http.StatusBadRequest)
						log.Error.Println("Error removing booked slot")
					}
					bookingIDtoRemove = 0
					prevUser.next = currentBooking.next
					b.size--
				}
			}

			//bookingIDtoRemove will set to zero when that booking is removed, if it is not set to zero after the whole loop, the booking could not be found
			if bookingIDtoRemove != 0 {
				http.Error(res, "No such booking/the booking has no longer exist!", http.StatusBadRequest)
				log.Warning.Println("Attempt to delete invalid booking ID")
				return
			}

		}
	}
	//template to obtain booking ID to be removed
	err := tpl.ExecuteTemplate(res, "removeBooking.gohtml", nil)
	if err != nil {
		log.Fatal.Fatalln(err)
	}

	currentBooking := b.head

	if currentBooking == nil {
		http.Error(res, "Empty Booking List", http.StatusBadRequest)
		log.Warning.Println("Empty booking list")
		return
	}
	//to loop through the linked list to display all the updated booking list
	for currentBooking != nil {
		err := tpl.ExecuteTemplate(res, "displayBooking.gohtml", currentBooking)
		if err != nil {
			log.Fatal.Fatalln(err)
		}
		currentBooking = currentBooking.next
	}

}

//BrowseBooking function will be triggered in two scenario. 
//1.Registered user can choose to browse their own booking record.
//2.Admin can browse the whole booking list. He/she can also filter the booking list by user name.
func (b *bookingList) BrowseBooking(res http.ResponseWriter, req *http.Request) {

	myUser := user.GetUser(res, req)

	var filterUser string

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			log.Panic.Println("Recovered from panic for browseBooking feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if !user.AlreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	if myUser.UserName == "admin" { //this portion is for browse booking list with admin login
		if req.Method == http.MethodPost {

			//getting admin choice of user
			filterUser = req.FormValue("filteruser")
		}

		err := tpl.ExecuteTemplate(res, "browseBooking.gohtml", user.MapUsers)
		if err != nil {
			log.Fatal.Fatalln(err)
		}
	} else { //this portion is built for browse my booking feature

		filterUser = myUser.UserName
		err := tpl.ExecuteTemplate(res, "browseMyBooking.gohtml", user.MapUsers)
		if err != nil {
			log.Fatal.Fatalln(err)
		}
	}

	currentBooking := b.head

	if currentBooking == nil {
		http.Error(res, "Empty Booking List", http.StatusBadRequest)
		log.Warning.Println("Empty booking list")
		return
	}

	for currentBooking != nil {
		if filterUser == "" { //this portion is for browse booking list with admin login
			err := tpl.ExecuteTemplate(res, "displayBooking.gohtml", currentBooking)
			if err != nil {
				log.Fatal.Fatalln(err)
			}
			currentBooking = currentBooking.next
		} else { //this portion is for browse booking list specfied user

			if currentBooking.Host == filterUser {
				err := tpl.ExecuteTemplate(res, "displayBooking.gohtml", currentBooking)
				if err != nil {
					log.Fatal.Fatalln(err)
				}

			}
			currentBooking = currentBooking.next
		}

	}

}
