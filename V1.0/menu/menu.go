//package menu.
package menu

import (
	"FunctionRoomsBookingSystem/booking"
	"FunctionRoomsBookingSystem/search"
	"FunctionRoomsBookingSystem/venue"
	"fmt"
)

//PrintMenu will display first page of the booking system.
func PrintMenu() {

	var choice int

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			fmt.Println("Oops, panic occurred:", err)
		}
	}()

	for choice != 6 {

		fmt.Printf("\n")
		fmt.Println("Welcome to Function Rooms Booking System")
		fmt.Println("==========================================")
		fmt.Println("1. Browse venue and availability")
		fmt.Println("2. Search for available venue")
		fmt.Println("3. Book Venue")
		fmt.Println("4. Remove Booking (for admin only)")
		fmt.Println("5. List whole booking list (for admin only)")
		fmt.Println("6. Exit the Booking System")
		fmt.Println("Select your choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			err := venue.ListRoom()
			if err != nil {
				fmt.Println(err)
			}
		case 2:
			_, _, _, err := search.EnterCriteria()
			if err != nil {
				fmt.Println(err)
			}
		case 3:
			bookings, err := booking.Create()
			if err != nil {
				fmt.Println(err)
			}
			if len(bookings) != 0 {
				err := venue.Room[bookings[0]%100-1].Slot.UpdateSlotAvailability(bookings)
				if err != nil {
					fmt.Println(err)
				}
			}
		case 4:
			fmt.Println("(Admin right is required to delete booking/view whole booking list)")
			err := booking.MainBookingList.RemoveBooking()
			if err != nil {
				fmt.Println(err)
			}

		case 5:
			fmt.Println("(Admin right is required to delete booking/view whole booking list)")
			err := booking.MainBookingList.FilterWholeBookingList()
			if err != nil {
				fmt.Println(err)
			}
		case 6:
			fmt.Println("Exiting the booking system")
		default:
			fmt.Println("Please select 1 to 6.")
		}

	}

}
