// want package:"def.GreekLetterChecked = {Alpha | Beta | Delta | Eta | Gamma}"
package def

import "fmt"

// checkenum
type GreekLetterChecked byte

const (
	Alpha GreekLetterChecked = iota
	Beta
	Gamma
	Delta
)

var Eta = GreekLetterChecked(5)

func NonExhaustiveList() {
	var x GreekLetterChecked = 99 // want "implicit conversion of 99 to def.GreekLetterChecked"
	x = 88                        // want "implicit conversion of 88 to def.GreekLetterChecked"
	switch x {                    // want "missing cases Delta, Eta and Gamma"
	case Alpha:
		fmt.Println("alpha")
	case Beta, 4: // want "implicit conversion of 4 to def.GreekLetterChecked"
		fmt.Println("beta")
	default: // want "def.GreekLetterChecked shouldn't have a default case"
		fmt.Println("default")
	}
}

type Struct struct {
	Value GreekLetterChecked
}

func AssignmentToStruct() {
	var s Struct

	s.Value = 123 // want "implicit conversion of 123 to def.GreekLetterChecked"
	s.Value = Alpha
}

func ExpandedAssignment() {
	var x GreekLetterChecked
	var s Struct

	s.Value, x = Values()
	_, _ = s.Value, x
}

func Values() (a, b GreekLetterChecked) {
	return Alpha, 3 // want "implicit conversion of 3 to def.GreekLetterChecked"
}

func ValuesX() (a, b GreekLetterChecked) {
	return Values()
}
