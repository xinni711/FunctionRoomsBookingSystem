//package search perform all the searching needed.
package search

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"FunctionRoomsBookingSystem/venue"
)

var (
	wg        sync.WaitGroup
	bookingID = []int{}
	err       error
)

//EnterCriteria to perform room search based on date, time, capacity and type.
func EnterCriteria() ([]int, int, int, error) {

	var dateS string
	var date, time, duration, participantSize, kind int
	totalDay := venue.DayOfMonth["February"]

	fmt.Println("Please enter your preferred date. (DDMM)")
	fmt.Scanln(&dateS)

	runes := []rune(dateS)  // this is to take care of condition like 0802, without trimming, it will have issue
	if runes[0] == '0' {
		strings.TrimLeft(dateS, "!0")
	}

	date, _ = strconv.Atoi(dateS)

	//error handling for date enter
	if (date%100) < 1 || (date%100) > 12 {
		return []int{}, 0, 0, errors.New("Invalid month, must be between 1 to 12")
	} else if (date % 100) != 2 {
		return []int{}, 0, 0, errors.New("Only February is open for booking now")
	} else if (date/100) < 1 || (date/100) > totalDay {
		return []int{}, 0, 0, errors.New("Invalid day entry")
	}

	fmt.Println("Please enter your preferred time. (eg.1000)")
	fmt.Scanln(&time)

	//error handling for time enter
	if time/100 < 0 || time/100 > 24 {
		return []int{}, 0, 0, errors.New("Invalid time")
	} else if time/100 < 10 || time/100 > 18 {
		return []int{}, 0, 0, errors.New("Outside of opening hours. Opening hours from 1000 to 1800")
	}

	fmt.Println("Please enter duration of the event in hours.")
	fmt.Scanln(&duration)

	//error handling for duration
	if duration > (1800 - time) {
		return []int{}, 0, 0, errors.New("Invalid duration, exceed opening hours")
	} else if duration < 1 || duration > 9 {
		return []int{}, 0, 0, errors.New("Invalid duration, duration has to be at least an hour and less than 9 hours")
	}

	fmt.Println("Please enter total number of participant.")
	fmt.Scanln(&participantSize)

	//error handling for size
	if participantSize < 1 {
		return []int{}, 0, 0, errors.New("Invalid participant size")
	} else if participantSize > 100 {
		return []int{}, 0, 0, errors.New("No suitable function room to fit the size of participants")
	}

	fmt.Println("Please enter the preferred type(1, 2 or 3). If there is none, press enter")
	fmt.Println("1. Meeting Room")
	fmt.Println("2. Activity Room")
	fmt.Println("3. Auditorium")
	fmt.Scanln(&kind)

	var kindFull string
	// error handling for kind, if not 1,2,3 invalid selection, press enter if there is no preferred type
	if kind < 0 || kind > 3 {
		return []int{}, 0, 0, errors.New("Invalid Selection")
	} else if kind == 0 {
		kindFull = "NoPreference"
	} else if kind == 1 {
		kindFull = "MeetingRoom"
	} else if kind == 2 {
		kindFull = "ActivityRoom"
	} else if kind == 3 {
		kindFull = "Auditorium"
	}

	bookingID, err = CheckCriteria(date, time, duration, participantSize, kindFull)

	return bookingID, duration, participantSize, err
}

//CheckCriteria based on criteria provided, search and sort available slot.
func CheckCriteria(date int, time int, duration int, size int, kind string) ([]int, error) {

	find := make(chan int)

	dateTime := (date/100)*10000 + time

	//fmt.Println(dateTime)

	wg.Add(len(venue.Room))

	for i := 1; i <= len(venue.Room); i++ {
		go ScreenSlot(find, (dateTime + i), duration)
	}

	slots := []int{}
	count := 0
	for count != len(venue.Room) {
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
		//fmt.Println(slots[i]%100 - 1)
		if venue.Room[slots[i]%100-1].Capacity >= size {
			if venue.Room[slots[i]%100-1].Kind == kind || kind == "NoPreference" {
				availableSlot = slots[i]
				bookingID = append(bookingID, availableSlot)
				//fmt.Println(bookingID)
				fmt.Printf("%d/%d %d:00 %s is available.\n", slots[i]/10000, date%100, (slots[i]%10000)/100, venue.Room[(slots[i]%100)-1].Name)

			}
		} else if venue.Room[slots[i]%100-1].Capacity < size && venue.Room[slots[i]%100-1].Kind == kind {
			return []int{}, errors.New("Meeting room cannot fit the size of participant")
		}
	}
	if len(bookingID) == 0 {
		return []int{}, errors.New("No available slots for preferred type of function room")
	}
	return bookingID, nil

}

//ScreenSlot was launched (10 goroutines) to perform slot availability binary search.
func ScreenSlot(find chan int, slotDetail int, duration int) error {

	defer wg.Done()
	i := slotDetail%100 - 1
	first := 0
	last := len(venue.Room[i].Slot) - 1
	wholeDuration := true
	for first <= last {
		mid := (first + last) / 2
		if venue.Room[i].Slot[mid].Info == slotDetail {
			for duration != 0 && wholeDuration != false { //if duration is 3 hours, this loop will loop 3 times to check if all 3 index is available
				wholeDuration = venue.Room[i].Slot[mid+duration-1].Available
				duration--
			}
			if wholeDuration == true {
				find <- venue.Room[i].Slot[mid].Info
			} else {
				find <- 0
			}
			return nil
		}

		if slotDetail < venue.Room[i].Slot[mid].Info {
			last = mid - 1
		} else {
			first = mid + 1
		}

	}
	return fmt.Errorf("Data not found")
}

// SelectionSort perform selection sort based on room number on the available slots array.
func selectionSort(arr []int, n int) {

	for last := n - 1; last >= 1; last-- {

		largest := indexOfLargest(arr, last+1)
		swap(&arr[largest], &arr[last])
	}
}

// indexOfLargest...
func indexOfLargest(arr []int, n int) int {
	largestIndex := 0
	for i := 1; i < n; i++ {
		if arr[i] > arr[largestIndex] {
			largestIndex = i
		}
	}
	return largestIndex
}

//Swap.
func swap(x *int, y *int) {
	temp := *x
	*x = *y
	*y = temp
}
