// want package:"def.Letter = {Alpha | Beta | Delta | Eta | Gamma}"
package def

import "fmt"

// Letter is an enumerated type.
type Letter byte // checkenum

const (
	Alpha Letter = iota
	Beta
	Gamma
	Delta
)

var Eta = Letter(5)

func NonExhaustiveList() {
	var x Letter = 99 // want "implicit conversion of 99 to def.Letter"
	x = 88            // want "implicit conversion of 88 to def.Letter"
	switch x {        // want "missing cases Delta, Eta and Gamma"
	case Alpha:
		fmt.Println("alpha")
	case Beta, 4: // want "implicit conversion of 4 to def.Letter"
		fmt.Println("beta")
	default: // want "def.Letter shouldn't have a default case"
		fmt.Println("default")
	}
}

type Struct struct {
	Value Letter
}

func AssignmentToStruct() {
	var s Struct

	s.Value = 123 // want "implicit conversion of 123 to def.Letter"
	s.Value = Alpha
}

func ExpandedAssignment() {
	var x Letter
	var s Struct

	s.Value, x = Values()
	_, _ = s.Value, x
}

func Values() (a, b Letter) {
	return Alpha, 3 // want "implicit conversion of 3 to def.Letter"
}

func ValuesX() (a, b Letter) {
	return Values()
}