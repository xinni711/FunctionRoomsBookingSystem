//Package booking stores all the booking variables and functions.
//It includes function to initialise preloaded bookings, performing venue booking search,
//and lastly generate booking record
package booking

import (
	"errors"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"

	"goInAction2Assignment/log"
	"goInAction2Assignment/user"
	"goInAction2Assignment/venue"
)

//booking struct is a type of every booking information.
type booking struct {
	BookingID       int64
	Venue           int
	BookingSlot     []int
	Duration        int
	participantSize int
	Host            string
	next            *booking
}

//bookingList struct is setup for linked list of booking.
type bookingList struct {
	head *booking
	size int
}

var (
	MainBookingList = &bookingList{nil, 0} //MainBookingList store all bookings in linked list.
	bookingCount    int64                  //bookingCount will be treated as booking ID. It is locked inside mutex so that there will be access once at a time.
	mu              sync.Mutex             //mu is used to lock the booking process where only one booking can be done at a time.
	tpl             *template.Template
	wg              sync.WaitGroup //wg is used to wait the go routine that being launched to preloaded some bookings.
	err             error
	mapBookingID    = map[string][]int{} //mapBookingID store the searchCriteria result from the booking query for every user.
)

func init() {

	//tpl read all the gohtml file in the templates folder.
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

}

//GenerateBookings initiate the application with some preloaded booking list
func GenerateBookings() error {

	runtime.GOMAXPROCS(2)
	bookingError := make(chan error, 6)

	wg.Add(6)

	//launch go routine to initiate some bookings at the start of the application, hence booking added is not in sequence
	go MainBookingList.addBooking(1, "leeken", []int{201401, 201501, 201601}, 3, 8, bookingError)
	go MainBookingList.addBooking(1, "tanalison", []int{171001, 171101}, 2, 6, bookingError)
	go MainBookingList.addBooking(3, "tanalison", []int{221203, 221303}, 2, 15, bookingError)
	go MainBookingList.addBooking(4, "ongryan", []int{31104, 31204, 31304, 31404}, 4, 20, bookingError)
	go MainBookingList.addBooking(6, "tanalison", []int{281006, 281106, 281206, 281306, 281406}, 5, 45, bookingError)
	go MainBookingList.addBooking(9, "limchristina", []int{181109, 181209, 181309}, 3, 70, bookingError)

	err := <-bookingError
	if err != nil {
		log.Error.Println("Booking initialization could not complete successfully, someone has booked it. Please try again.")
		return err
	}

	wg.Wait()

	log.Info.Println("Bookings preloaded done!")

	return nil
}

//BookVenue function will first perform a search based on the booking input by user.
//After that, all available venue that matches the criteria will be listed out. Registered user can select the venue and hit the submit button to book it.
func BookVenue(res http.ResponseWriter, req *http.Request) {

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			log.Panic.Println("Recover from panic for bookVenue feature:", err)
			return
		}
	}()

	if !user.AlreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	var roomtoBook, duration, participantSize int
	totalDay := venue.DayOfMonth["March"]
	myUser := user.GetUser(res, req)
	if req.Method == http.MethodPost {
		// get form values
		dateS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("date")))
		regexdateS := regexp.MustCompile(`^[0-3][0-9](0|1)[0-9]$`)
		if !regexdateS.MatchString(dateS) {
			http.Error(res, "Invalid date selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Booking, invalid date selection detected")
			return
		}
		runes := []rune(dateS) // this is to take care of condition like 0802, without trimming, it will have issue
		if runes[0] == '0' {
			dateS = strings.TrimLeft(dateS, "!0")
		}
		date, _ := strconv.Atoi(dateS)

		timeS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("time")))
		regextimeS := regexp.MustCompile(`^[0-1][0-9]00$`)
		if !regextimeS.MatchString(timeS) {
			http.Error(res, "Invalid time selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Booking, invalid time selection detected")
			return
		}
		time, _ := strconv.Atoi(timeS)

		durationS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("duration")))
		regexdurationS := regexp.MustCompile(`^[0-9]$`)
		if !regexdurationS.MatchString(durationS) {
			http.Error(res, "Invalid duration selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Booking, invalid duration selection detected")
			return
		}
		duration, _ = strconv.Atoi(durationS)

		participantsizeS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("participantSize")))
		regexparticipantsizeS := regexp.MustCompile(`^[\d]{1,3}$`)
		if !regexparticipantsizeS.MatchString(participantsizeS) {
			http.Error(res, "Invalid participant size input, please try again", http.StatusBadRequest)
			log.Info.Println("Booking, invalid participant size input detected")
			return
		}
		participantSize, _ = strconv.Atoi(participantsizeS)

		kindS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("kind")))
		regexkindS := regexp.MustCompile(`^(1|2|3?)$`)
		if !regexkindS.MatchString(kindS) {
			http.Error(res, "Invalid function room type selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Booking, invalid function room type selection detected")
			return
		}
		kind, _ := strconv.Atoi(kindS)

		//error handling for date enter
		if (date%100) < 1 || (date%100) > 12 {
			http.Error(res, "Invalid month, must be between 1 to 12", http.StatusBadRequest)
			log.Info.Println("Booking, date entry error detected")
			return
		} else if (date % 100) != 3 {
			http.Error(res, "Only March is open for booking now", http.StatusBadRequest)
			log.Info.Println("Booking, date entry error detected")
			return
		} else if (date/100) < 1 || (date/100) > totalDay {
			http.Error(res, "Invalid day entry", http.StatusBadRequest)
			log.Info.Println("Booking, date entry error detected")
			return
		}

		//error handling for time enter
		if time/100 < 0 || time/100 > 24 {
			http.Error(res, "Invalid time", http.StatusBadRequest)
			log.Info.Println("Booking, time entry error detected")
			return
		} else if time/100 < 10 || time/100 > 18 {
			http.Error(res, "Outside of opening hours. Opening hours from 1000 to 1800", http.StatusBadRequest)
			log.Info.Println("Booking, time entry error detected")
			return
		}

		//error handling for duration
		if duration > (1800 - time) {
			http.Error(res, "Invalid duration, exceed opening hours", http.StatusBadRequest)
			log.Info.Println("Booking, duration entry error detected")
			return
		} else if duration < 1 || duration > 9 {
			http.Error(res, "Invalid duration, duration has to be at least an hour and less than 9 hours", http.StatusBadRequest)
			log.Info.Println("Booking, duration entry error detected")
			return
		}

		//error handling for participant size
		if participantSize < 1 {
			http.Error(res, "Invalid participant size", http.StatusBadRequest)
			log.Info.Println("Booking, participant size entry error detected")
			return
		} else if participantSize > 100 {
			http.Error(res, "No suitable function rooms to fit the size of participants", http.StatusBadRequest)
			log.Info.Println("Booking, participant size entry error detected")
			return
		}

		var kindFull string
		// error handling for kind, if not 1,2,3 invalid selection, press enter if there is no preferred type
		if kind < 0 || kind > 3 {
			http.Error(res, "Invalid type selection, please select only 1,2 or 3", http.StatusBadRequest)
			log.Info.Println("Booking, function room type entry error detected")
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

		//fmt.Println("Checking criteria")
		mapBookingID[myUser.UserName], err = venue.CheckCriteria(date, time, duration, participantSize, kindFull)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			log.Info.Println(err.Error())
			return
		}

		for i := 0; i < len(mapBookingID[myUser.UserName]); i++ {
			venue.AvailableVmap[myUser.UserName] = append(venue.AvailableVmap[myUser.UserName], venue.VenueDetail{mapBookingID[myUser.UserName][i] / 10000, 3, (mapBookingID[myUser.UserName][i] % 10000) / 100, mapBookingID[myUser.UserName][i] % 100, duration, participantSize})

		}

		//log.Info.Println("Available slots: ", venue.AvailableVmap[myUser.UserName], ", booking query by", myUser.UserName)

	}

	//this one to cater for the second submit button in the page
	if req.Method == http.MethodGet && venue.AvailableVmap[myUser.UserName] != nil {

		roomtobookS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("roomtoBook")))
		regexroomtobookS := regexp.MustCompile(`^(1?)[0-9]$`)
		if !regexroomtobookS.MatchString(roomtobookS) {
			http.Error(res, "Invalid room selection, please try again.", http.StatusBadRequest)
			log.Warning.Println("Booking, invalid choice of room to book detected")
			return
		}
		roomtoBook, _ = strconv.Atoi(roomtobookS)

		var selected int
		var bookings []int

		for i := 0; i < len(mapBookingID[myUser.UserName]); i++ {
			if mapBookingID[myUser.UserName][i]%100 == roomtoBook {
				selected = mapBookingID[myUser.UserName][i]
				break
			}
		}

		// create an array based list of booking details for the whole duration, if duration is 3 hours, array will have 3 inputs
		totalDuration := venue.AvailableVmap[myUser.UserName][0].Duration
		for totalDuration != 0 {
			bookings = append(bookings, selected)
			selected += 100
			totalDuration--
		}

		//log.Info.Println("Selected room", roomtoBook, ", Slots to book:", bookings)
		bookingError := make(chan error)
		wg.Add(1)

		go MainBookingList.addBooking(roomtoBook, myUser.UserName, bookings, venue.AvailableVmap[myUser.UserName][0].Duration, venue.AvailableVmap[myUser.UserName][0].Participant, bookingError)
		err := <-bookingError
		if err != nil {
			http.Error(res, "Booking requested could not complete successfully, someone has booked it. Please try to book another slot.", http.StatusBadRequest)
			log.Info.Println("Unsuccessful booking, booking was done by other user.", err, myUser.UserName)
			return
		}

		wg.Wait()

		//pop up message to confirm booking
		err = tpl.ExecuteTemplate(res, "bookConfirmation.gohtml", roomtoBook)
		if err != nil {
			log.Fatal.Fatalln(err)
		}

		log.Info.Println("Booking is confirmed. Venue booked is", roomtoBook, "Slots booked:", bookings)

		_, ok := venue.AvailableVmap[myUser.UserName]
		if ok {
			delete(venue.AvailableVmap, myUser.UserName)
			roomtoBook = 0
			mapBookingID[myUser.UserName] = nil
		}

	}

	err1 := tpl.ExecuteTemplate(res, "bookVenue.gohtml", venue.AvailableVmap[myUser.UserName])
	if err1 != nil {
		log.Fatal.Fatalln(err)
	}

}

//addBooking function traverse through booking linked list to add new booking record.
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

	err := venue.Room[option-1].Slot.UpdateSlotAvailability(bookings)
	if err != nil {
		currentBooking.next = nil
		log.Info.Println("Booking requested could not complete successfully, someone has booked it. Booking requested:", bookings, "by", userName)
		bookingError <- errors.New("booking requested could not complete successfully, someone has booked it")
		return err
	}
	mu.Unlock()

	bookingError <- err

	return nil
}
