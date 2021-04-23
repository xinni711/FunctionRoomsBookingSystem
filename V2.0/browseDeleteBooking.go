package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// booking will be removed based on booking ID. User can select the booking ID based on the summary of booking displayed
func (b *bookingList) removeBooking(res http.ResponseWriter, req *http.Request) {

	var bookingIDToRemove int64

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			fmt.Println("Oops, panic occurred for removeBooking feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if req.Method == http.MethodPost {

		//obtain user input of booking ID to remove
		bookingIDToRemove, _ = strconv.ParseInt(req.FormValue("bookingIDToDelete"), 10, 64)  
		
		if bookingIDToRemove != 0 {
			
			currentBooking := b.head

			if b.head == nil {
				http.Error(res, "Empty booking list, nothing to be deleted", http.StatusBadRequest)
				return
			}

			if currentBooking.next == nil {
				if currentBooking.BookingID == bookingIDToRemove {
					//remove booked slot, reset them back to available
					fmt.Println("Deleting booking ID", currentBooking.BookingID)
					room[currentBooking.BookingSlot[0]%100-1].Slot.RemoveBookedSlot(currentBooking.BookingSlot)
					b.head = nil
				}
			}

			for i := 1; i < b.size; i++ {
				prevUser := currentBooking
				currentBooking = currentBooking.next
				if currentBooking.BookingID == bookingIDToRemove {
					fmt.Println("Deleting booking ID", currentBooking.BookingID)
					room[currentBooking.BookingSlot[0]%100-1].Slot.RemoveBookedSlot(currentBooking.BookingSlot)
					bookingIDToRemove = 0
					prevUser.next = currentBooking.next
					b.size--
				}
			}

			//bookingIDToRemove will set to zero when that booking is removed, if it is not set to zero after the whole loop, the booking could not be found
			if bookingIDToRemove !=0{
				http.Error(res, "No such booking/the booking has no longer exist!", http.StatusBadRequest)
				return
			}
			
		}
	}
	//template to obtain booking ID to be removed
	err := tpl.ExecuteTemplate(res, "removeBooking.gohtml", nil)
	if err != nil {
		log.Fatalln(err)
	}

	currentBooking := b.head
	
	if currentBooking == nil {
		http.Error(res, "Empty Booking List", http.StatusBadRequest)
		return
	}
	//to loop through the linked list to display all the updated booking list
	for currentBooking != nil {
		err := tpl.ExecuteTemplate(res, "displayBookingcopy.gohtml", currentBooking)
		if err != nil {
			log.Fatalln(err)
		}
		currentBooking = currentBooking.next
	}

}

// RemovedBookedSlot help to update the slot back to true when the admin remove the slot.
func (s slotArr) RemoveBookedSlot(bookings []int) error {

	duration := len(bookings)
	first := 0
	last := len(s)
	for first <= last {
		mid := (first + last) / 2
		
		if s[mid].Info == bookings[0] {
			for duration != 0 { //if duration is 3 hours, this loop will loop 3 times to update
				(&s[mid+duration-1]).Available = true
				duration--
			}
			fmt.Println("Removed Booked Slot")
			fmt.Println(bookings)
			return nil
		}

		if bookings[0] < s[mid].Info {
			last = mid - 1
		} else {
			first = mid + 1
		}
		
	}
	return errors.New("Slot Not found")
}


func (b *bookingList) browseBooking(res http.ResponseWriter, req *http.Request) {

	myUser := getUser(res, req)

	var filterUser string

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			fmt.Println("Oops, panic occurred for browseBooking feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if myUser.UserName == "admin"{ //this portion is for browse booking list with admin login
		if req.Method == http.MethodPost {

			//getting admin choice of user
			filterUser = req.FormValue("filteruser")
		}
	
		err := tpl.ExecuteTemplate(res, "browseBooking.gohtml", mapUsers)
		if err != nil {
			log.Fatalln(err)
		}
	} else { //this portion is built for browse my booking feature

		filterUser = myUser.UserName
		err := tpl.ExecuteTemplate(res, "browseMyBooking.gohtml", mapUsers)
		if err != nil {
			log.Fatalln(err)
		}
	}
	
	
	currentBooking := b.head

	if currentBooking == nil {
		http.Error(res, "Empty Booking List", http.StatusBadRequest)
		return
	}

	for currentBooking != nil {
		if filterUser == "" { //this portion is for browse booking list with admin login
			err := tpl.ExecuteTemplate(res, "displayBooking.gohtml", currentBooking)
			if err != nil {
				log.Fatalln(err)
			}
			currentBooking = currentBooking.next
		} else { //this portion is for browse booking list specfied user

			if currentBooking.Host == filterUser {
				err := tpl.ExecuteTemplate(res, "displayBooking.gohtml", currentBooking)
				if err != nil {
					log.Fatalln(err)
				}
				
			}
			currentBooking = currentBooking.next
		}

	}

	
	

}
