// want package:"defs.Day = {Friday | Monday | Saturday | Sunday | Thursday | Tuesday | Wednesday}"
package defs

import "fmt"

// Day is an enumerated type.
type Day string // checkenum

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

	switch day { // want "missing cases Friday, Monday, Saturday, Sunday, Thursday and Wednesday"
	case "monday": // want "implicit conversion of \"monday\" to defs.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	default: // want "defs.Day shouldn't have a default case"
		fmt.Println("default")
	}
}
