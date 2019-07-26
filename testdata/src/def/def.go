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

func Example() {
	var x GreekLetterChecked = 99 // want "basic literal declaration to checked enum"
	x = 88                        // want "basic literal assignment to checked enum"
	switch x {                    // want "switch clause missing for Delta, Eta and Gamma"
	case Alpha:
		fmt.Println("alpha")
	case Beta, 4: // want "basic literal clause for checked enum"
		fmt.Println("beta")
	default: // want "default literal clause for checked enum"
		fmt.Println("default")
	}
}
