// want package:"enumstruct.Day = {False | Maybe | True}"
package enumstruct

import "fmt"

// Option is an enumerated type.
type Option struct{ value string } // enumcheck

var (
	True  = Option{"true"}
	False = Option{"false"}
	Maybe = Option{"maybe"}
)

func DayNonExhaustive() {
	var day Option

	switch day { // want "missing cases False and Maybe"
	case Option{"invalid"}: // want "invalid enum for enumstruct.Option"
		fmt.Println("beta")
	case True:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}
