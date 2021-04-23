package main

import (
	"log"
	"net/http"
	"strconv"
)

//Slot is setup with 3 information.
type Slot struct {
	Info      int
	Available bool
}

type slotArr []Slot

//RoomInfo equipped with basic info of room and also an array of its slot availability
type roomInfo struct {
	Name     string
	Kind     string
	Capacity int
	Slot     slotArr
}

//Rooms available to pick from are declared upfront
var (
	room = []roomInfo{
		{Name: "MR01", Kind: "MeetingRoom", Capacity: 10, Slot: TotalAvailableSlot[:279]},
		{Name: "MR02", Kind: "MeetingRoom", Capacity: 10, Slot: TotalAvailableSlot[279:558]},
		{Name: "MR03", Kind: "MeetingRoom", Capacity: 10, Slot: TotalAvailableSlot[558:837]},
		{Name: "MR04", Kind: "MeetingRoom", Capacity: 20, Slot: TotalAvailableSlot[837:1116]},
		{Name: "MR05", Kind: "MeetingRoom", Capacity: 20, Slot: TotalAvailableSlot[1116:1395]},
		{Name: "AR06", Kind: "ActivityRoom", Capacity: 50, Slot: TotalAvailableSlot[1395:1674]},
		{Name: "AR07", Kind: "ActivityRoom", Capacity: 50, Slot: TotalAvailableSlot[1674:1953]},
		{Name: "AR08", Kind: "ActivityRoom", Capacity: 50, Slot: TotalAvailableSlot[1953:2232]},
		{Name: "AD09", Kind: "Auditorium", Capacity: 100, Slot: TotalAvailableSlot[2232:2511]},
		{Name: "AD10", Kind: "Auditorium", Capacity: 100, Slot: TotalAvailableSlot[2511:2790]},
	}
	availableSlot      = make(slotArr, 0, 279)
	TotalAvailableSlot = make(slotArr, 0, 2790)
)

//DayOfMonth assign days to different months
//Not really useful at this stage of project,setup upfront for possible future expansion
var dayOfMonth = map[string]int{
	"January": 31, "February": 28, "March": 31, "April": 30, "May": 31, "June": 30,
	"July": 31, "August": 31, "September": 30, "October": 31, "November": 30, "December": 31,
}

//generateSlots initiate slice of slots based on date, month and venue (XXXXXX)
func generateSlots() {

	//this loop generate 4 digit slot info for availableSlot array for a room, eg: 0314 --> March 03 14:00
	for i := 1; i <= dayOfMonth["March"]; i++ {
		for j := 10; j <= 18; j++ {
			slotDetail := Slot{(i * 100) + j, true}
			availableSlot = append(availableSlot, slotDetail)
		}
	}

	//this one combine availableSlot for all rooms together --279 slot per room (10 rooms)
	//the slot generation was seperated into two loops to avoid O(n3) complexity
	for r := 0; r < len(room); r++ {
		for a := 0; a < 279; a++ {
			sum := (availableSlot[a].Info * 100) + (r + 1)
			slotWithRoomDetail := Slot{sum, true}
			TotalAvailableSlot = append(TotalAvailableSlot, slotWithRoomDetail)
		}

	}

}

//show table of slot based on user venue selection
func browseVenue(res http.ResponseWriter, req *http.Request) {

	var roomChoice string
	if req.Method == http.MethodPost {

		roomChoice = req.FormValue("roomChoice")
		roomChoiceI, _ := strconv.Atoi(roomChoice)
		err := tpl.ExecuteTemplate(res, "displayVenueAvailability.gohtml", roomChoiceI)
		if err != nil {
			log.Fatalln(err)
		}
		printTableOfSlot(res, req, roomChoiceI-1)

	}

	if roomChoice == "" {
		err := tpl.ExecuteTemplate(res, "browseVenue.gohtml", room)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

//print out the availability of all slots for a particular venue
func printTableOfSlot(res http.ResponseWriter, req *http.Request, num int) {

	var slotToPrint []Slot

	for j := 0; j < (len(room[num].Slot)); j++ {

		slotToPrint = append(slotToPrint, room[num].Slot[j])

		if (j+1)%9 == 0 {
			data := struct {
				Day         int
				SlotToPrint []Slot
			}{(j + 1) / 9, slotToPrint}

			err := tpl.ExecuteTemplate(res, "displayVenueDayAvailability.gohtml", data)
			if err != nil {
				log.Fatalln(err)
			}
			slotToPrint = []Slot{}

		}
	}
}

