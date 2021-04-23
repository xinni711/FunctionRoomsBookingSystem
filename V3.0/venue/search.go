package venue

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"goInAction2Assignment/log"
	"goInAction2Assignment/user"
)

var (
	wg           sync.WaitGroup
	err          error
	mapBookingID = map[string][]int{} //mapBookingID store the results from checkCriteria function.
)

//VenueDetails is a data type that store details of user search criteria of venue.
type VenueDetail struct {
	Day         int
	Month       int
	Time        int
	Room        int
	Duration    int
	Participant int
}

//AvailableVmap store all the search results for a particular user.
var AvailableVmap = map[string][]VenueDetail{}

//SearchVenue function enable the user to key in all their criteria in the browser and perform a search on available venue based on the inputs.
func SearchVenue(res http.ResponseWriter, req *http.Request) {
	myCookie, _ := req.Cookie("myCookie")
	totalDay := DayOfMonth["March"]

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			log.Panic.Println("Recovered from panic for searchVenue feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if req.Method == http.MethodPost {
		// get form values
		dateS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("date")))
		regexdateS := regexp.MustCompile(`^[0-3][0-9](0|1)[0-9]$`)
		if !regexdateS.MatchString(dateS) {
			http.Error(res, "Invalid date selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Invalid date selection detected")
			return
		}
		runes := []rune(dateS) // this is to take care of condition like 0803, without trimming, it will have issue
		if runes[0] == '0' {
			dateS = strings.TrimLeft(dateS, "!0")
		}
		date, _ := strconv.Atoi(dateS)

		timeS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("time")))
		regextimeS := regexp.MustCompile(`^[0-1][0-9]00$`)
		if !regextimeS.MatchString(timeS) {
			http.Error(res, "Invalid time selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Invalid time selection detected")
			return
		}
		time, _ := strconv.Atoi(timeS)

		durationS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("duration")))
		regexdurationS := regexp.MustCompile(`^[0-9]$`)
		if !regexdurationS.MatchString(durationS) {
			http.Error(res, "Invalid duration input, please try again.", http.StatusBadRequest)
			log.Info.Println("Search venue, invalid duration input detected")
			return
		}
		duration, _ := strconv.Atoi(durationS)

		participantsizeS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("participantSize")))
		regexparticipantsizeS := regexp.MustCompile(`^[\d]{1,3}$`)
		if !regexparticipantsizeS.MatchString(participantsizeS) {
			http.Error(res, "Invalid participant size input, please try again", http.StatusBadRequest)
			log.Info.Println("Search venue, invalid participant size input detected")
			return
		}
		participantSize, _ := strconv.Atoi(participantsizeS)

		kindS := user.Policy.Sanitize(strings.TrimSpace(req.FormValue("kind")))
		regexkindS := regexp.MustCompile(`^(1|2|3?)$`)
		if !regexkindS.MatchString(kindS) {
			http.Error(res, "Invalid function room type selection, please try again.", http.StatusBadRequest)
			log.Info.Println("Search venue, invalid function room type selection detected")
			return
		}
		kind, _ := strconv.Atoi(kindS)

		//error handling for date enter
		if (date%100) < 1 || (date%100) > 12 {
			http.Error(res, "Invalid month, must be between 1 to 12", http.StatusBadRequest)
			log.Info.Println("Search venue, date entry error detected")
			return
		} else if (date % 100) != 3 {
			http.Error(res, "Only March is open for booking now", http.StatusBadRequest)
			log.Info.Println("Search venue, date entry error detected")
			return
		} else if (date/100) < 1 || (date/100) > totalDay {
			http.Error(res, "Invalid day entry", http.StatusBadRequest)
			log.Info.Println("Search venue, date entry error detected")
			return
		}

		//error handling for time enter
		if time/100 < 0 || time/100 > 24 {
			http.Error(res, "Invalid time", http.StatusBadRequest)
			log.Info.Println("Search venue, time entry error detected")
			return
		} else if time/100 < 10 || time/100 > 18 {
			http.Error(res, "Outside of opening hours. Opening hours from 1000 to 1800", http.StatusBadRequest)
			log.Info.Println("Search venue, time entry error detected")
			return
		}

		//error handling for duration
		if duration > (1900 - time) {
			http.Error(res, "Invalid duration, exceed opening hours", http.StatusBadRequest)
			log.Info.Println("Search venue, duration entry error detected")
			return
		} else if duration < 1 || duration > 9 {
			http.Error(res, "Invalid duration, duration has to be at least an hour and less than 9 hours", http.StatusBadRequest)
			log.Info.Println("Search venue, duration entry error detected")
			return
		}

		//error handling for size
		if participantSize < 1 {
			http.Error(res, "Invalid participant size", http.StatusBadRequest)
			log.Info.Println("Search venue, participant size entry error detected")
			return
		} else if participantSize > 100 {
			http.Error(res, "No suitable function room to fit the size of participants", http.StatusBadRequest)
			log.Info.Println("Search venue, participant size entry error detected")
			return
		}

		var kindFull string
		// error handling for kind, if not 1,2,3 invalid selection, press enter if there is no preferred type
		if kind < 0 || kind > 3 {
			http.Error(res, "Invalid type selection", http.StatusBadRequest)
			log.Info.Println("Search venue, function room type entry error detected")
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

		mapBookingID[myCookie.Value], err = CheckCriteria(date, time, duration, participantSize, kindFull)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			log.Info.Println("Check criteria-", err)
			return
		}

		for i := 0; i < len(mapBookingID[myCookie.Value]); i++ {
			//map cookie value to search result to present concurrency issue.
			AvailableVmap[myCookie.Value] = append(AvailableVmap[myCookie.Value], VenueDetail{mapBookingID[myCookie.Value][i] / 10000, date % 100, (mapBookingID[myCookie.Value][i] % 10000) / 100, mapBookingID[myCookie.Value][i] % 100, duration, participantSize})
		}

		//log.Info.Println("Search result for is", AvailableVmap[myCookie.Value])

	}

	err := tpl.ExecuteTemplate(res, "searchVenue.gohtml", AvailableVmap[myCookie.Value])
	if err != nil {
		log.Fatal.Fatalln(err)
	}

	_, ok := AvailableVmap[myCookie.Value]
	if ok {
		delete(AvailableVmap, myCookie.Value)
	}

}

//CheckCriteria based on criteria provided, search and sort available slot
func CheckCriteria(date int, time int, duration int, size int, kind string) ([]int, error) {

	find := make(chan int)

	dateTime := (date/100)*10000 + time

	wg.Add(len(Room))

	for i := 1; i <= len(Room); i++ {
		go screenSlot(find, (dateTime + i), duration)
	}

	slots := []int{}
	count := 0
	for count != len(Room) {
		slot := <-find
		if slot != 0 {
			slots = append(slots, slot)
		}
		count++
	}

	wg.Wait()

	selectionSort(slots, len(slots))

	var availableSlot int
	bookingID := []int{}
	for i := 0; i < len(slots); i++ {
		if Room[slots[i]%100-1].Capacity >= size {
			if Room[slots[i]%100-1].Kind == kind || kind == "NoPreference" {
				availableSlot = slots[i]
				bookingID = append(bookingID, availableSlot)
				//fmt.Printf("%d/%d %d:00 %s is available.\n", slots[i]/10000, date%100, (slots[i]%10000)/100, room[(slots[i]%100)-1].Name)

			}
		} else if Room[slots[i]%100-1].Capacity < size && Room[slots[i]%100-1].Kind == kind {
			return []int{}, errors.New("meeting room cannot fit the size of participant")
		}
	}
	if len(bookingID) == 0 {
		return []int{}, errors.New("no available slots for preferred type of function room")
	}
	return bookingID, nil
}

//ScreenSlot was launched (10 goroutines) to perform slot availability binary search
func screenSlot(find chan int, slotDetail int, duration int) error {

	defer wg.Done()
	i := slotDetail%100 - 1
	first := 0
	last := len(Room[i].Slot) - 1
	wholeDuration := true

	for first <= last {
		mid := (first + last) / 2
		if Room[i].Slot[mid].Info == slotDetail {
			for duration != 0 && wholeDuration { //if duration is 3 hours, this loop will loop 3 times to check if all 3 index is available
				wholeDuration = Room[i].Slot[mid+duration-1].Available
				duration--
			}
			if wholeDuration {
				find <- Room[i].Slot[mid].Info
			} else {
				find <- 0
			}
			return nil
		}

		if slotDetail < Room[i].Slot[mid].Info {
			last = mid - 1
		} else {
			first = mid + 1
		}
	}
	return fmt.Errorf("data not found")
}

//Selection sort was performed based on room number on the available slots array.
func selectionSort(arr []int, n int) {

	for last := n - 1; last >= 1; last-- {

		largest := indexOfLargest(arr, last+1)
		swap(&arr[largest], &arr[last])
	}
}

//indexOfLargest return the largest number in the slice.
func indexOfLargest(arr []int, n int) int {
	largestIndex := 0
	for i := 1; i < n; i++ {
		if arr[i] > arr[largestIndex] {
			largestIndex = i
		}
	}
	return largestIndex
}

//swap perform a swap in location between two variable.
func swap(x *int, y *int) {
	temp := *x
	*x = *y
	*y = temp
}
