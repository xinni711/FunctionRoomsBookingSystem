package booking

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

//user information are stored in struct form.
type user struct {
	firstName string
	lastName  string
	id        string
	password  string
	category  int
	next      *user
}

//user are stored in linked list.
type userList struct {
	head *user
	size int
}

var (
	mainList = &userList{nil, 0}
	reader   = bufio.NewReader(os.Stdin)
)

//GenerateUser initiate preliminary user list.
func GenerateUser() {

	mainList.addUser("Matthew", "Lee", "lee.matthew", "leematthew", 2)
	mainList.addUser("Alison", "Tan", "tan.alison", "tanalison", 1)
	mainList.addUser("Christina", "Lim", "lim.christina", "limchristina", 1)
	mainList.addUser("Ryan", "Ong", "ong.ryan", "ongryan", 1)

}

// To register as new user before booking if the person is not registered user.
func register() (string, error) {

	fmt.Println("What is your first name?")
	firstName, _ := reader.ReadString('\n')
	firstName = strings.TrimRight(firstName, "\n")

	if firstName == "" {
		return "", errors.New("Invalid input")
	}

	fmt.Println("What is your last name?")
	lastName, _ := reader.ReadString('\n')
	lastName = strings.TrimRight(lastName, "\n")

	if lastName == "" {
		return "", errors.New("Invalid input")
	}

	fmt.Println("What is your preferred userID?")
	id, _ := reader.ReadString('\n')
	id = strings.TrimRight(id, "\n")

	if len(id) < 6 {
		return "", errors.New("Id has to be more than 6 characters")
	}

	fmt.Println("Please enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimRight(id, "\n")

	if len(password) < 6 {
		return "", errors.New("Password has to be more than 6 charactwers")
	}

	mainList.addUser(firstName, lastName, id, password, 1)

	return id, nil

}

// after getting input from register function, will run this function to add user to the user list.
func (u *userList) addUser(fName string, lName string, userid string, pass string, cat int) error {
	newUser := &user{
		firstName: fName,
		lastName:  lName,
		id:        userid,
		password:  pass,
		category:  cat,
		next:      nil,
	}

	if u.head == nil {
		u.head = newUser
	} else {
		currentUser := u.head
		for currentUser.next != nil {
			currentUser = currentUser.next
		}
		currentUser.next = newUser
	}
	u.size++

	return nil

}

//CheckAuthentication will be prompt to user to login before booking venue, edit booking or delete booking.
func CheckAuthentication() (string, int, error) {

	var option int

	fmt.Println("Are you registered user?")
	fmt.Println("Select 1 for registered user, select 2 to register new user")
	fmt.Scanln(&option)

	//error handling for option selection
	if option < 1 || option > 2 {
		return "", 0, errors.New("Invalid option")
	} else if option == 2 {
		userID, err := register()
		if err != nil {
			return "", 0, err
		}
		fmt.Println("New user registered!")
		return userID, 1, nil
	} else if option == 1 {
		fmt.Println("Please enter your userID")
		userID, _ := reader.ReadString('\n')
		userID = strings.TrimRight(userID, "\n")

		fmt.Println("Please enter your password")
		password, _ := reader.ReadString('\n')
		password = strings.TrimRight(password, "\n")

		fmt.Println("Hold on, logging in......")
		time.Sleep(500 * time.Millisecond)

		//traverse through the list to check if existing userID else show error
		//if is registered user, check if password and userID matched, if not return return error
		valid, cat, err := mainList.checkUser(userID, password)
		if valid == true {
			fmt.Println("Successful Login!")
			return userID, cat, nil
		}
		return "", 0, err
	}
	return "", 0, nil
}

//not needed for the application, but can be add as additional feature
/* func (u *userList) printWholeList() error {
	currentUser := u.head
	if currentUser == nil {
		fmt.Println("User list is empty.")
		return nil
	}

	fmt.Printf(currentUser.id)

	for currentUser.next != nil {
		currentUser = currentUser.next
		fmt.Printf(currentUser.id)
	}

	return nil
} */

//check if userID and password key in by registered user is matched.
func (u *userList) checkUser(userID string, password string) (bool, int, error) {
	currentUser := u.head
	if currentUser == nil {
		return false, 0, errors.New("No such user")
	}
	for i := 1; i <= u.size; i++ {
		if currentUser.id == userID && currentUser.password == password {
			return true, currentUser.category, nil
		} else if currentUser.id == userID && currentUser.password != password {
			return false, currentUser.category, errors.New("Mismatch userID and password")
		}
		currentUser = currentUser.next
	}
	return false, 0, errors.New("no such User, please register account first")
}
