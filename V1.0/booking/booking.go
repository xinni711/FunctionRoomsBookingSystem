//package booking settle all the booking.
package booking

import (
	"FunctionRoomsBookingSystem/search"
	"FunctionRoomsBookingSystem/venue"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

// booking struct of every booking information.
type booking struct { 
	bookingID       int64
	venue           int
	bookingSlot     []int
	duration        int
	participantSize int
	host            string
	next            *booking
}

//bookingList linked list struct for booking.
type bookingList struct { 
	head *booking
	size int
}

var (
	
	MainBookingList = &bookingList{nil, 0} //MainBookingList store all bookings in linked list.
	bookingCount    int64
	wg              sync.WaitGroup
)

//GenerateBookings initiate the application with some preloaded booking list.
func GenerateBookings() error {

	runtime.GOMAXPROCS(2)
	wg.Add(6)

	go MainBookingList.addBooking(1, "lee.matthew", []int{201401, 201501, 201601}, 3, 8)
	go MainBookingList.addBooking(1, "tan.alison", []int{171001, 171101}, 2, 6)
	go MainBookingList.addBooking(3, "tan.alison", []int{221203, 221303}, 2, 15)
	go MainBookingList.addBooking(4, "ong.ryan", []int{31104, 31204, 31304, 31404}, 4, 20)
	go MainBookingList.addBooking(6, "tan.alison", []int{281006, 281106, 281206, 281306, 281406}, 5, 45)
	go MainBookingList.addBooking(9, "lim.christina", []int{181109, 181209, 181309}, 3, 70)
	err := venue.Room[201401%100-1].Slot.UpdateSlotAvailability([]int{201401, 201501, 201601})
	err = venue.Room[171001%100-1].Slot.UpdateSlotAvailability([]int{171001, 171101})
	err = venue.Room[221203%100-1].Slot.UpdateSlotAvailability([]int{221203, 221303})
	err = venue.Room[31104%100-1].Slot.UpdateSlotAvailability([]int{31104, 31204, 31304, 31404})
	err = venue.Room[280906%100-1].Slot.UpdateSlotAvailability([]int{281006, 281106, 281206, 281306, 281406})
	err = venue.Room[181109%100-1].Slot.UpdateSlotAvailability([]int{181109, 181209, 181309})

	wg.Wait()

	if err != nil {
		return err
	}

	return nil
}

//Create booking function is used to obtain all the required information to create booking.
func Create() ([]int, error) {

	var option int

	fmt.Println("Login is needed for venue booking")
	userID, _, err := CheckAuthentication()
	if err != nil {
		return []int{}, err
	}

	fmt.Println("Hello", userID, "!")

	bookingID, duration, participantSize, err := search.EnterCriteria()
	if err != nil {
		return []int{}, err
	}

	fmt.Println("Select your venue based on availability")
	fmt.Scanln(&option)

	if option < 1 || option > 10 {
		return []int{}, errors.New("Invalid Function Room Selection")
	}

	var selected int
	var bookings []int
	for i := 0; i < len(bookingID); i++ {
		if bookingID[i]%100 == option {
			selected = bookingID[i]
			continue
		}
		if i == len(bookingID) && bookingID[i]%100 != option {
			return []int{}, errors.New("Room selected not available")
		}

	}

	// create an array based list of booking details for the whole duration, if duration is 3 hours, array will have 3 inputs.
	totalDuration := duration
	for totalDuration != 0 {
		bookings = append(bookings, selected)
		selected += 100
		totalDuration--
	}

	wg.Add(1)
	go MainBookingList.addBooking(option, userID, bookings, duration, participantSize)

	wg.Wait()
	fmt.Println("Booking done!")
	return bookings, nil
}

// addBooking traverse booking linked list to add new booking.
func (b *bookingList) addBooking(option int, userID string, bookings []int, duration int, size int) error {

	defer wg.Done()

	newBooking := &booking{
		bookingID:       bookingCount,
		venue:           option,
		bookingSlot:     bookings,
		duration:        duration,
		participantSize: size,
		host:            userID,
		next:            nil,
	}

	atomic.LoadInt64(&bookingCount)

	if b.head == nil {
		b.head = newBooking
	} else {
		currentBooking := b.head
		for currentBooking.next != nil {
			currentBooking = currentBooking.next
		}
		currentBooking.next = newBooking
	}
	b.size++
	atomic.AddInt64(&bookingCount, 1)
	return nil

}

//PrintWholeBookingList enabled admin to view whole booking list.
func (b *bookingList) PrintWholeBookingList(choice int, user string) error {

	userID, cat, err := CheckAuthentication()
	if err != nil {
		return err
	}

	if cat != 2 {
		return errors.New("sorry, remove booking/print whole booking list are not allowed. You are not an admin")
	}

	fmt.Printf("Welcome %s\n", userID)
	fmt.Printf("Displaying booking list.....\n")

	currentBooking := b.head
	if currentBooking == nil {
		fmt.Println("Booking list is empty.")
		return nil
	}

	//currentBooking.printBooking() //print booking details
	var count int
	for currentBooking != nil {
		if choice == 1 {
			currentBooking.printBooking() //print booking details
		} else if choice == 2 && currentBooking.host == user {
			currentBooking.printBooking()
			count++
		}
		currentBooking = currentBooking.next
	}

	if choice == 2 && count == 0 {
		return errors.New("No booking found for this userID")
	}

	return nil
}

//printbooking.....
func (b *booking) printBooking() {
	fmt.Printf("Booking ID: %+v, Bookings for %v hours: %v , Host: %v, Venue: %v\n",
		b.bookingID, b.duration, b.bookingSlot, b.host, b.venue)
}

//Remove booking will remove particular booking from booking list by entering booking ID. However, remaining booking ID will not be updated.
func (b *bookingList) RemoveBooking() error {

	var id int64
	err := MainBookingList.PrintWholeBookingList(1, "")
	if err != nil {
		return err
	}
	fmt.Println("Which booking ID you wish to remove?")
	fmt.Scanln(&id)

	currentBooking := b.head

	if b.head == nil {
		return errors.New("empty booking list")
	}

	if currentBooking.next == nil {
		if currentBooking.bookingID == id {
			//remove booked slot
			venue.Room[currentBooking.bookingSlot[0]%100-1].Slot.RemoveBookedSlot(currentBooking.bookingSlot)
			b.head = nil
			return nil
		}
		return errors.New("no such booking/The booking has no longer exist")
	}

	for i := 1; i < b.size; i++ {
		prevUser := currentBooking
		currentBooking = currentBooking.next
		if currentBooking.bookingID == id {
			venue.Room[currentBooking.bookingSlot[0]%100-1].Slot.RemoveBookedSlot(currentBooking.bookingSlot)
			fmt.Println("Deleting booking ID", currentBooking.bookingID)
			prevUser.next = currentBooking.next
			return nil
		}
	}

	return errors.New("no such booking/the booking has no longer exist")

}

//filterWholeBookingList.....
func (b *bookingList) FilterWholeBookingList() error {

	var choice int
	var userID string
	fmt.Println("Which booking list you would like to view?")
	fmt.Println("1. Complete Booking List (based on booking ID)")
	fmt.Println("2. Booking List of particular user")
	fmt.Scanln(&choice)

	if choice == 2 {
		fmt.Println("Which user ID you would like to view?")
		fmt.Scanln(&userID)
	}

	err := MainBookingList.PrintWholeBookingList(choice, userID)
	if err != nil {
		return err
	}
	return nil
}
