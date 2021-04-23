//Package venue store information of all the function rooms available. It also store the slot availability of a venue in hour.
//User of the application can view the venue availability using the functions in this package.
//Any modification of the availability of the slot can be performed using the function in this package as well.
package venue

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"goInAction2Assignment/log"
	"goInAction2Assignment/user"
)

//Each slot is setup with 2 information, slot details and its availability.
type Slot struct {
	Info      int
	Available bool
}

//All the slots are store in slice of struct data type.
type SlotArr []Slot

//RoomInfo equipped with basic info of room and also an array of its slot availability.
type roomInfo struct {
	Name     string
	Kind     string
	Capacity int
	Slot     SlotArr
}

var tpl *template.Template

//Rooms available to pick from are stored in slice of struct data type.
//This slice contains all the details information of the function room such as name, kind, capacity and slot availability.
var (
	Room = []roomInfo{
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
	availableSlot      = make(SlotArr, 0, 279)
	TotalAvailableSlot = make(SlotArr, 0, 2790)
)

//DayOfMonth assign days to different months
//Not really useful at this stage of project,setup upfront for possible future expansion
var DayOfMonth = map[string]int{
	"January": 31, "February": 28, "March": 31, "April": 30, "May": 31, "June": 30,
	"July": 31, "August": 31, "September": 30, "October": 31, "November": 30, "December": 31,
}

func init() {

	//read all the gohtml file in the templates folder
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

}

//GenerateSlots initiate slice of slots based on date, month and venue (XXXXXX)
func GenerateSlots() {

	//this loop generate 4 digit slot info for availableSlot array for a room, eg: 0314 --> March 03 14:00
	for i := 1; i <= DayOfMonth["March"]; i++ {
		for j := 10; j <= 18; j++ {
			slotDetail := Slot{(i * 100) + j, true}
			availableSlot = append(availableSlot, slotDetail)
		}
	}

	//this one combine availableSlot for all rooms together --279 slot per room (10 rooms)
	//the slot generation was seperated into two loops to avoid O(n3) complexity
	for r := 0; r < len(Room); r++ {
		for a := 0; a < 279; a++ {
			sum := (availableSlot[a].Info * 100) + (r + 1)
			slotWithRoomDetail := Slot{sum, true}
			TotalAvailableSlot = append(TotalAvailableSlot, slotWithRoomDetail)
		}

	}

}

//BrowseVenue function show table of slot availabiltiy based on user venue selection.
func BrowseVenue(res http.ResponseWriter, req *http.Request) {

	defer func() { //to handle any possible panic
		if err := recover(); err != nil {
			log.Panic.Println("Recovered from panic for browseVenue feature:", err)
			http.Error(res, "500 - Something bad happened:", http.StatusInternalServerError)
			return
		}
	}()

	var roomChoice string
	
	if req.Method == http.MethodPost {

		roomChoice = user.Policy.Sanitize(strings.TrimSpace(req.FormValue("roomChoice")))
		regexproomChoice := regexp.MustCompile(`^[\d]{1,2}$`)
		if !regexproomChoice.MatchString(roomChoice) {
			http.Error(res, "Invalid selection", http.StatusBadRequest)
			log.Warning.Println("Invalid roomChoice selection detected")
			return
		}

		roomChoiceI, _ := strconv.Atoi(roomChoice)
		err1 := tpl.ExecuteTemplate(res, "displayVenueAvailability.gohtml", roomChoiceI)
		if err1 != nil {
			log.Fatal.Fatalln(err)
		}
		printTableOfSlot(res, req, roomChoiceI-1)

	}

	if roomChoice == "" {
		err := tpl.ExecuteTemplate(res, "browseVenue.gohtml", Room)
		if err != nil {
			log.Fatal.Fatalln(err)
		}
	}
}

//printTableOfSlot print out the availability of all slots for a particular venue.
func printTableOfSlot(res http.ResponseWriter, req *http.Request, num int) {

	var slotToPrint []Slot

	for j := 0; j < (len(Room[num].Slot)); j++ {

		slotToPrint = append(slotToPrint, Room[num].Slot[j])

		if (j+1)%9 == 0 {
			data := struct {
				Day         int
				SlotToPrint []Slot
			}{(j + 1) / 9, slotToPrint}

			err := tpl.ExecuteTemplate(res, "displayVenueDayAvailability.gohtml", data)
			if err != nil {
				log.Fatal.Fatalln(err)
			}
			slotToPrint = []Slot{}

		}
	}
}
