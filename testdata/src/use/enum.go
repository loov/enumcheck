package use

import (
	"fmt"

	"def"
)

func NonExhaustiveList() {
	var x def.Letter = 99 // want "implicit conversion of 99 to def.Letter"
	x = 88                // want "implicit conversion of 88 to def.Letter"
	switch x {            // want "missing cases Delta, Eta and Gamma"
	case def.Alpha:
		fmt.Println("alpha")
	case def.Beta, 4: // want "implicit conversion of 4 to def.Letter"
		fmt.Println("beta")
	default: // want "def.Letter shouldn't have a default case"
		fmt.Println("default")
	}
}
