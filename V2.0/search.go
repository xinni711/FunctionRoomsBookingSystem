package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	wg        sync.WaitGroup
	bookingID = []int{}
	err       error
)

type venueDetail struct {
	Day         int
	Month       int
	Time        int
	Room        int
	Duration    int
	Participant int
}

var availableV []venueDetail
var availableVmap = map[string][]venueDetail{}

func searchVenue(res http.ResponseWriter, req *http.Request) {
	myCookie, _ := req.Cookie("myCookie")
	totalDay := dayOfMonth["March"]

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			fmt.Println("Oops, panic occurred for searchVenue feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	if req.Method == http.MethodPost {
		// get form values
		time, _ := strconv.Atoi(req.FormValue("time"))
		dateS := req.FormValue("date")
		duration, _ := strconv.Atoi(req.FormValue("duration"))
		participantSize, _ := strconv.Atoi(req.FormValue("participantSize"))
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

		//error handling for size
		if participantSize < 1 {
			http.Error(res, "Invalid participant size", http.StatusBadRequest)
			return
		} else if participantSize > 100 {
			http.Error(res, "No suitable function room to fit the size of participants", http.StatusBadRequest)
			return
		}

		var kindFull string
		// error handling for kind, if not 1,2,3 invalid selection, press enter if there is no preferred type
		if kind < 0 || kind > 3 {
			http.Error(res, "Invalid type selection", http.StatusBadRequest)
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

		bookingID, err = checkCriteria(date, time, duration, participantSize, kindFull)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		for i := 0; i < len(bookingID); i++ {
			//map cookie value to search result to present concurrency issue.
			availableVmap[myCookie.Value] = append(availableVmap[myCookie.Value], venueDetail{bookingID[i] / 10000, date % 100, (bookingID[i] % 10000) / 100, bookingID[i] % 100, duration, participantSize})
		}

		fmt.Println("Search result for", myCookie.Value, "is", availableVmap[myCookie.Value])

	}

	err := tpl.ExecuteTemplate(res, "searchVenue.gohtml", availableVmap[myCookie.Value])
	if err != nil {
		log.Fatalln(err)
	}

	_, ok := availableVmap[myCookie.Value]
	if ok {
		delete(availableVmap, myCookie.Value)
	}

}

//CheckCriteria based on criteria provided, search and sort available slot
func checkCriteria(date int, time int, duration int, size int, kind string) ([]int, error) {

	find := make(chan int)

	dateTime := (date/100)*10000 + time

	wg.Add(len(room))

	for i := 1; i <= len(room); i++ {
		go screenSlot(find, (dateTime + i), duration)
	}

	slots := []int{}
	count := 0
	for count != len(room) {
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
		if room[slots[i]%100-1].Capacity >= size {
			if room[slots[i]%100-1].Kind == kind || kind == "NoPreference" {
				availableSlot = slots[i]
				bookingID = append(bookingID, availableSlot)
				//fmt.Printf("%d/%d %d:00 %s is available.\n", slots[i]/10000, date%100, (slots[i]%10000)/100, room[(slots[i]%100)-1].Name)

			}
		} else if room[slots[i]%100-1].Capacity < size && room[slots[i]%100-1].Kind == kind {
			return []int{}, errors.New("Meeting room cannot fit the size of participant")
		}
	}
	if len(bookingID) == 0 {
		return []int{}, errors.New("No available slots for preferred type of function room")
	}
	return bookingID, nil
}

//ScreenSlot was launched (10 goroutines) to perform slot availability binary search
func screenSlot(find chan int, slotDetail int, duration int) error {

	defer wg.Done()
	i := slotDetail%100 - 1
	first := 0
	last := len(room[i].Slot) - 1
	wholeDuration := true

	for first <= last {
		mid := (first + last) / 2
		if room[i].Slot[mid].Info == slotDetail {
			for duration != 0 && wholeDuration != false { //if duration is 3 hours, this loop will loop 3 times to check if all 3 index is available
				wholeDuration = room[i].Slot[mid+duration-1].Available
				duration--
			}
			if wholeDuration == true {
				find <- room[i].Slot[mid].Info
			} else {
				find <- 0
			}
			return nil
		}

		if slotDetail < room[i].Slot[mid].Info {
			last = mid - 1
		} else {
			first = mid + 1
		}
	}
	return fmt.Errorf("Data not found")
}

// perform selection sort based on room number on the available slots array
func selectionSort(arr []int, n int) {

	for last := n - 1; last >= 1; last-- {

		largest := indexOfLargest(arr, last+1)
		swap(&arr[largest], &arr[last])
	}
}

func indexOfLargest(arr []int, n int) int {
	largestIndex := 0
	for i := 1; i < n; i++ {
		if arr[i] > arr[largestIndex] {
			largestIndex = i
		}
	}
	return largestIndex
}

func swap(x *int, y *int) {
	temp := *x
	*x = *y
	*y = temp
}
