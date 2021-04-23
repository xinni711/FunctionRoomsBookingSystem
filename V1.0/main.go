//package main is XXXX.
package main

import (
	"FunctionRoomsBookingSystem/menu"
	"FunctionRoomsBookingSystem/venue"
	"FunctionRoomsBookingSystem/booking"
)

//init to intialise.
func init() {
	venue.GenerateSlots()
	booking.GenerateUser()
	booking.GenerateBookings()
}

//main start with printmenu.
func main() {

	menu.PrintMenu()

}
