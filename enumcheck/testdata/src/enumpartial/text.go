// want package:"enumpartial.Day = {Friday | Monday | Saturday | Sunday | Thursday | Tuesday | Wednesday}"
package enumpartial

import "fmt"

// Day is an enumerated type.
//enumcheck:silent
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

func DayNonExhaustive() {
	var day Day

	switch day { //enumcheck:exhaustive // want "missing cases Friday, Monday, Saturday, Sunday, Thursday and Wednesday"
	case "monday": // want "implicit conversion of \"monday\" to enumpartial.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}

func DayBasic() {
	var day Day

	switch day {
	case "monday": // want "implicit conversion of \"monday\" to enumpartial.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}
