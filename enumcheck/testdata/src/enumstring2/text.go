// want package:"enumstring2.Day = {Friday | Monday | Saturday | Sunday | Thursday | Tuesday | Wednesday}"
package enumstring2

import "fmt"

// Day is an enumerated type.
//
//enumcheck:exhaustive
type Day string

const (
	Monday    Day = "monday"
	Tuesday   Day = "tuesday"
	Wednesday Day = "wednesday"
	Thursday  Day = "thursday"
	Friday    Day = "friday"
	Saturday  Day = "saturday"
)

const Sunday Day = "sunday"

func DayNonExhaustive() {
	var day Day

	switch day { // want "missing cases Friday, Monday, Saturday, Sunday, Thursday and Wednesday"
	case "monday": // want "implicit conversion of \"monday\" to enumstring2.Day"
		fmt.Println("monday")
	case Tuesday:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}
