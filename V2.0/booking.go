package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type booking struct { // struct of every booking information
	BookingID       int64
	Venue           int
	BookingSlot     []int
	Duration        int
	participantSize int
	Host            string
	next            *booking
}

type bookingList struct { //linked list struct for booking
	head *booking
	size int
}

var (
	//MainBookingList store all bookings in linked list
	MainBookingList = &bookingList{nil, 0}
	bookingCount    int64
	mu              sync.Mutex   // locking the booking process where only one booking can be done at a time
)

//generateBookings initiate the application with some preloaded booking list
func generateBookings() error {

	runtime.GOMAXPROCS(2)
	bookingError:=make(chan error, 6)
	
	wg.Add(6)

	//launch go routine to initiate some bookings at the start of the application, hence booking added is not in sequence
	go MainBookingList.addBooking(1, "lee.matthew", []int{201401, 201501, 201601}, 3, 8, bookingError)
	go MainBookingList.addBooking(1, "tan.alison", []int{171001, 171101}, 2, 6, bookingError)
	go MainBookingList.addBooking(3, "tan.alison", []int{221203, 221303}, 2, 15, bookingError)
	go MainBookingList.addBooking(4, "ong.ryan", []int{31104, 31204, 31304, 31404}, 4, 20, bookingError)
	go MainBookingList.addBooking(6, "tan.alison", []int{281006, 281106, 281206, 281306, 281406}, 5, 45, bookingError)
	go MainBookingList.addBooking(9, "lim.christina", []int{181109, 181209, 181309}, 3, 70, bookingError)
	
	err = <-bookingError
	if err!=nil{
		fmt.Println("Booking requested could not complete successfully, someone has booked it. Please try again.")
		return err
	}
	
	wg.Wait()
	
	fmt.Println("Bookings preloaded done!")

	return nil
}


func bookVenue(res http.ResponseWriter, req *http.Request) {

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			fmt.Println("Oops, panic occurred for bookVenue feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	var roomToBook, duration, participantSize int
	totalDay := dayOfMonth["March"]
	myUser := getUser(res, req)
	if req.Method == http.MethodPost {
		// get form values
		time, _ := strconv.Atoi(req.FormValue("time"))
		dateS := req.FormValue("date")
		duration, _ = strconv.Atoi(req.FormValue("duration"))
		participantSize, _ = strconv.Atoi(req.FormValue("participantSize"))
		kind, _ := strconv.Atoi(req.FormValue("kind"))

		runes := []rune(dateS) // this is to take care of condition like 0802, without trimming, it will have issue
		if runes[0] == '0' {
			strings.TrimLeft(dateS, "!0")
		}

		date, _ := strconv.Atoi(dateS)

		//error handling for date enter
		if (date%100) < 1 || (date%100) > 12 {
			http.Error(res, "Invalid month, must be between 1 to 12", http.StatusBadRequest)
			return
		} else if (date % 100) != 3 {
			http.Error(res, "Only March is open for booking now", http.StatusBadRequest)
			return
		} else if (date/100) < 1 || (date/100) > totalDay {
			http.Error(res, "Invalid day entry", http.StatusBadRequest)
			return
		}

		//error handling for time enter
		if time/100 < 0 || time/100 > 24 {
			http.Error(res, "Invalid time", http.StatusBadRequest)
			return
		} else if time/100 < 10 || time/100 > 18 {
			http.Error(res, "Outside of opening hours. Opening hours from 1000 to 1800", http.StatusBadRequest)
			return
		}

		//error handling for duration
		if duration > (1800 - time) {
			http.Error(res, "Invalid duration, exceed opening hours", http.StatusBadRequest)
			return
		} else if duration < 1 || duration > 9 {
			http.Error(res, "Invalid duration, duration has to be at least an hour and less than 9 hours", http.StatusBadRequest)
			return
		}

		//error handling for participant size
		if participantSize < 1 {
			http.Error(res, "Invalid participant size", http.StatusBadRequest)
			return
		} else if participantSize > 100 {
			http.Error(res, "No suitable function rooms to fit the size of participants", http.StatusBadRequest)
			return
		}

		var kindFull string
		// error handling for kind, if not 1,2,3 invalid selection, press enter if there is no preferred type
		if kind < 0 || kind > 3 {
			http.Error(res, "Invalid type selection, please select only 1,2 or 3", http.StatusBadRequest)
			return
		} else if kind == 0 {
			kindFull = "NoPreference"
		} else if kind == 1 {
			kindFull = "MeetingRoom"
		} else if kind == 2 {
			kindFull = "ActivityRoom"
		} else if kind == 3 {
			kindFull = "Auditorium"
		}
		
		fmt.Println("Checking criteria")
		bookingID, err = checkCriteria(date, time, duration, participantSize, kindFull)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		
		for i := 0; i < len(bookingID); i++ {
			availableVmap[myUser.UserName] = append(availableVmap[myUser.UserName], venueDetail{bookingID[i] / 10000, 3, (bookingID[i] % 10000) / 100, bookingID[i] % 100, duration, participantSize})
			
		}

		fmt.Println("Available slots: ",availableVmap[myUser.UserName], ", booking query by",myUser.UserName)
	}

	//this one to cater for the second submit button in the page
	if req.Method == http.MethodGet && availableVmap[myUser.UserName] != nil {

		roomToBook, _ = strconv.Atoi(req.FormValue("roomToBook"))

		var selected int
		var bookings []int
		for i := 0; i < len(bookingID); i++ {
			if bookingID[i]%100 == roomToBook {
				selected = bookingID[i]
				break
			}
		}

		// create an array based list of booking details for the whole duration, if duration is 3 hours, array will have 3 inputs
		totalDuration := availableVmap[myUser.UserName][0].Duration
		for totalDuration != 0 {
			bookings = append(bookings, selected)
			selected += 100
			totalDuration--
		}

		fmt.Println("Selected room", roomToBook,", Slots to book:", bookings)
		bookingError:=make(chan error)
		wg.Add(1)
		
		go MainBookingList.addBooking(roomToBook, myUser.UserName, bookings, availableVmap[myUser.UserName][0].Duration, availableVmap[myUser.UserName][0].Participant, bookingError)
		err:=<-bookingError
		if err!=nil{
			http.Error(res, "Booking requested could not complete successfully, someone has booked it. Please try to book another slot.", http.StatusBadRequest)
			return
		} 	

		wg.Wait()

	
		//pop up message to confirm booking
		err = tpl.ExecuteTemplate(res, "bookConfirmation.gohtml", roomToBook)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("Booking for", myUser.UserName, "is confirmed. Venue booked is", roomToBook)

		_, ok := availableVmap[myUser.UserName];
    	if ok {
        	delete(availableVmap, myUser.UserName);
   	 	}

	}

	err := tpl.ExecuteTemplate(res, "bookVenue.gohtml", availableVmap[myUser.UserName])
	if err != nil {
		log.Fatalln(err)
	}

}

// traverse booking linked list to add new booking
func (b *bookingList) addBooking(option int, userName string, bookings []int, duration int, size int, bookingError chan error) error {

	defer wg.Done()
	mu.Lock()
	newBooking := &booking{
		BookingID:       bookingCount,
		Venue:           option,
		BookingSlot:     bookings,
		Duration:        duration,
		participantSize: size,
		Host:            userName,
		next:            nil,
	}

	atomic.LoadInt64(&bookingCount)
	var currentBooking = b.head
	if b.head == nil {
		b.head = newBooking
	} else {
		//currentBooking := b.head
		for currentBooking.next != nil {
			currentBooking = currentBooking.next
		}
		currentBooking.next = newBooking
	}
	b.size++
	atomic.AddInt64(&bookingCount, 1)

	err := room[option-1].Slot.updateSlotAvailability(bookings)
	if err != nil {
		currentBooking.next=nil
		fmt.Fprintln(os.Stderr, "Booking requested could not complete successfully, someone has booked it. Booking requested:", bookings, "by",userName)
		bookingError <- errors.New("booking requested could not complete successfully, someone has booked it")
		return err
	}
	mu.Unlock()

	bookingError <- err
	
	return nil
}

//UpdateSlotAvailability help to update the slot to false when the venue is booked.
func (s slotArr) updateSlotAvailability(bookings []int) error {

	duration := len(bookings)
	first := 0
	last := len(s)
	for first <= last {
		mid := (first + last) / 2

		if s[mid].Info == bookings[0] {
			for duration != 0 { //if duration is 3 hours, this loop will loop 3 times to update
				if (&s[mid+duration-1]).Available == true {  //double checking if the search result is still valid
					(&s[mid+duration-1]).Available = false
					fmt.Println("Booking", s[mid+duration-1].Info, ", change availability to", s[mid+duration-1].Available)
					duration--
				} else {
					// if one of the slot is detected to be booked, the previous few slots that has been booked will be removed and return booking of this event is not successful
					toBeRevert := len(bookings) - duration  
					for toBeRevert != 0 {
						(&s[mid+len(bookings)-toBeRevert]).Available = true
						fmt.Println("Unsuccessful Booking", s[mid+len(bookings)-toBeRevert].Info, ", change availability back to", s[mid+len(bookings)-toBeRevert].Info)
						toBeRevert--
					}
					return errors.New("the slots have been booked")
				}

			}
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
