package venue

import (
	"errors"
	"goInAction2Assignment/log"
)

//RemovedBookedSlot help to update the slot back to true when the admin remove the slot.
func (s SlotArr) RemoveBookedSlot(bookings []int) error {

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
			log.Info.Println("Removed Booked Slot", bookings)
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

//UpdateSlotAvailability help to update the slot to false when the venue is booked.
func (s SlotArr) UpdateSlotAvailability(bookings []int) error {

	duration := len(bookings)
	first := 0
	last := len(s)
	for first <= last {
		mid := (first + last) / 2

		if s[mid].Info == bookings[0] {
			for duration != 0 { //if duration is 3 hours, this loop will loop 3 times to update
				if (&s[mid+duration-1]).Available { //double checking if the search result is still valid
					(&s[mid+duration-1]).Available = false
					//log.Info.Println("Booking", s[mid+duration-1].Info, ", change availability to", s[mid+duration-1].Available)
					duration--
				} else {
					// if one of the slot is detected to be booked, the previous few slots that has been booked will be removed and return booking of this event is not successful
					toBeRevert := len(bookings) - duration
					for toBeRevert != 0 {
						(&s[mid+len(bookings)-toBeRevert]).Available = true
						log.Info.Println("Unsuccessful Booking", s[mid+len(bookings)-toBeRevert].Info, ", change availability back to", s[mid+len(bookings)-toBeRevert].Info)
						toBeRevert--
					}
					return errors.New("the slots have been booked")
				}

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
