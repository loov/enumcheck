// want package:"defs.DayChecked = {Friday | Monday | Saturday | Sunday | Thursday | Tuesday | Wednesday}"
package defs

import "fmt"

// checkenum
type DayChecked string

const (
	Monday    = DayChecked("monday")
	Tuesday   = DayChecked("tuesday")
	Wednesday = DayChecked("wednesday")
	Thursday  = DayChecked("thursday")
	Friday    = DayChecked("friday")
	Saturday  = DayChecked("saturday")
	Sunday    = DayChecked("sunday")
)

func DayNonExhaustive() {
	var day DayChecked

	switch day { // want "missing cases Friday, Monday, Saturday, Sunday, Thursday and Wednesday"
	case "monday": // want "implicit conversion of \"monday\" to defs.DayChecked"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	default: // want "defs.DayChecked shouldn't have a default case"
		fmt.Println("default")
	}
}
