//package venue.
package venue

import (
	"errors"
	"fmt"
)

//Slot is setup with 3 information.
type Slot struct {
	Info      int
	Available bool
}

//slotArr.
type slotArr []Slot

//Rooms available to pick from are declared upfront.
var (
	Room = []RoomInfo{
		{Name: "MR01", Kind: "MeetingRoom", Capacity: 10, Slot: TotalAvailableSlot[:252]},
		{Name: "MR02", Kind: "MeetingRoom", Capacity: 10, Slot: TotalAvailableSlot[252:504]},
		{Name: "MR03", Kind: "MeetingRoom", Capacity: 10, Slot: TotalAvailableSlot[504:756]},
		{Name: "MR04", Kind: "MeetingRoom", Capacity: 20, Slot: TotalAvailableSlot[756:1008]},
		{Name: "MR05", Kind: "MeetingRoom", Capacity: 20, Slot: TotalAvailableSlot[1008:1260]},
		{Name: "AR06", Kind: "ActivityRoom", Capacity: 50, Slot: TotalAvailableSlot[1260:1512]},
		{Name: "AR07", Kind: "ActivityRoom", Capacity: 50, Slot: TotalAvailableSlot[1512:1764]},
		{Name: "AR08", Kind: "ActivityRoom", Capacity: 50, Slot: TotalAvailableSlot[1764:2016]},
		{Name: "AD09", Kind: "Auditorium", Capacity: 100, Slot: TotalAvailableSlot[2016:2268]},
		{Name: "AD10", Kind: "Auditorium", Capacity: 100, Slot: TotalAvailableSlot[2268:2520]},
	}
	availableSlot      = make(slotArr, 0, 252)
	TotalAvailableSlot = make(slotArr, 0, 2520)
)

//RoomInfo equipped with basic info of room and also an array of its slot availability
type RoomInfo struct {
	Name     string
	Kind     string
	Capacity int
	Slot     slotArr
}

//ListRoom print all the basic information of the room
//it also let the user decide if they want to view the availability of the slot in that room
func ListRoom() error {

	var choice int
	var num int
	PrintRoom()
	fmt.Println("Select 1 to check venue availability")
	fmt.Println("Select 2 to return")
	fmt.Scanln(&choice)

	if choice == 1 {
		fmt.Println("Which venue to view?")
		fmt.Scanln(&num)
		printTableOfSlot(num - 1)
		return nil
	} else if choice == 2 {
		return nil
	} else {
		return errors.New("Invalid selection")
	}

}

//PrintRoom list down details of rooms.
func PrintRoom() {
	fmt.Println("List of venue: ")
	for i := 0; i < len(Room); i++ {
		fmt.Printf("Room: %d --> Name: %s , Type: %s, Capacity: %d \n", i+1, Room[i].Name, Room[i].Kind, Room[i].Capacity)
	}
}

//printTableofSlot this function print the availablility slot of a room in table form.
func printTableOfSlot(num int) {

	var toPrint int
	fmt.Println("     Available Slots For Room ", num+1)
	fmt.Println("=======================================")
	fmt.Println("     10  11  12  13  14  15  16  17  18")

	for j := 0; j < (len(Room[num].Slot)); j++ {
		if Room[num].Slot[j].Available == true {
			toPrint = 1
		} else {
			toPrint = 0
		}
		if (j)%9 == 0 {
			if Room[num].Slot[j].Info/10000 < 10 {
				fmt.Printf(" %d    ", Room[num].Slot[j].Info/10000)
			} else {
				fmt.Printf("%d    ", Room[num].Slot[j].Info/10000)
			}
		}
		fmt.Printf("%d   ", toPrint)
		if (j+1)%9 == 0 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

//DayOfMonth assign days to different months.
//Not really useful at this stage of project,setup upfront for possible future expansion.
var DayOfMonth = map[string]int{
	"January": 31, "February": 28, "March": 31, "April": 30, "May": 31, "June": 30,
	"July": 31, "August": 31, "September": 30, "October": 31, "November": 30, "December": 31,
}

//GenerateSlots initiate slice of slots based on date, month and venue (XXXXXX).
func GenerateSlots() {

	//this loop generate 4 digit slot info for availableSlot array for a room, eg: 0214 --> February 02 14:00
	for i := 1; i <= DayOfMonth["February"]; i++ {
		for j := 10; j <= 18; j++ {
			slotDetail := Slot{(i * 100) + j, true}
			availableSlot = append(availableSlot, slotDetail)
		}
	}

	//this one combine availableSlot for all rooms together --252 slot per room (10 rooms)
	//the slot generation was seperated into two loops to avoid O(n3) complexity
	for r := 0; r < len(Room); r++ {
		for a := 0; a < 252; a++ {
			sum := (availableSlot[a].Info * 100) + (r + 1)
			slotWithRoomDetail := Slot{sum, true}
			TotalAvailableSlot = append(TotalAvailableSlot, slotWithRoomDetail)
		}

	}

}

//UpdateSlotAvailability help to update the slot to false when the venue is booked.
func (s slotArr) UpdateSlotAvailability(bookings []int) error {

	duration := len(bookings)
	first := 0
	last := len(s)
	for first <= last {
		mid := (first + last) / 2
		if s[mid].Info == bookings[0] {
			for duration != 0 { //if duration is 3 hours, this loop will loop 3 times to update
				(&s[mid+duration-1]).Available = false
				//fmt.Println("Booking", s[mid+duration-1].Info, ", change availability to", s[mid+duration-1].Available)
				duration--
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
