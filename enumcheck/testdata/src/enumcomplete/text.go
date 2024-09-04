// want package:"enumcomplete.Day = {Friday | Monday | Saturday | Sunday | Thursday | Tuesday | Wednesday}"
package enumcomplete

import "fmt"

// Day is an enumerated type.
//
//enumcheck:complete
type Day string

const (
	Monday    = Day("monday")
	Tuesday   = Day("tuesday")
	Wednesday = Day("wednesday")
	Thursday  = Day("thursday")
	Friday    = Day("friday")
	Saturday  = Day("saturday")
	Sunday    = Day("sunday")
)

func DayWithDefault() {
	var day Day

	switch day {
	case "monday": // want "implicit conversion of \"monday\" to enumcomplete.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}

func DayWithoutDefault() {
	var day Day

	switch day { // want "missing cases Friday, Monday, Saturday, Sunday, Thursday and Wednesday"
	case "monday": // want "implicit conversion of \"monday\" to enumcomplete.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	}
}

func DayBasic() {
	var day Day

	switch day { // want "missing cases Friday, Monday, Saturday, Sunday, Thursday and Wednesday"
	case "monday": // want "implicit conversion of \"monday\" to enumcomplete.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	}
}
