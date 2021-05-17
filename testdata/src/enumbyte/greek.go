// want package:"enumbyte.Letter = {Alpha | Beta | Delta | Eta | Gamma}"
package enumbyte

import "fmt"

// Letter is an enumerated type.
type Letter byte // enumcheck

const (
	Alpha Letter = iota
	Beta
	Gamma
	Delta
)

var Eta = Letter(5)

func NonExhaustiveList() {
	var x Letter = 99 // want "implicit conversion of 99 to enumbyte.Letter"
	x = 88            // want "implicit conversion of 88 to enumbyte.Letter"
	switch x {        // want "missing cases Delta, Eta and Gamma"
	case Alpha:
		fmt.Println("alpha")
	case Beta, 4: // want "implicit conversion of 4 to enumbyte.Letter"
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}

func ToString(v Letter) string { return string(v) }

func ImplicitConversion() {
	var _ Letter = 90 // want "implicit conversion of 90 to enumbyte.Letter"
	_ = ToString(80)  // want "implicit conversion of 80 to enumbyte.Letter"
	ToString(70)      // want "implicit conversion of 70 to enumbyte.Letter"
}

type Struct struct {
	Value Letter
}

func AssignmentToStruct() {
	var s Struct

	s.Value = 123 // want "implicit conversion of 123 to enumbyte.Letter"
	s.Value = Alpha
}

func ExpandedAssignment() {
	var x Letter
	var s Struct

	s.Value, x = Values()
	_, _ = s.Value, x
}

func Values() (a, b Letter) {
	return Alpha, 3 // want "implicit conversion of 3 to enumbyte.Letter"
}

func ValuesX() (a, b Letter) {
	return Values()
}

func Chan() {
	ch := make(chan Letter, 10)
	ch <- 123 // want "implicit conversion of 123 to enumbyte.Letter"
}

func NamedChan() {
	type LetterChan chan Letter
	ch := make(LetterChan)
	ch <- 123 // want "implicit conversion of 123 to enumbyte.Letter"
}

func ChanFunc() {
	fn := func() chan Letter { return make(chan Letter, 10) }
	fn() <- 123 // want "implicit conversion of 123 to enumbyte.Letter"
}
