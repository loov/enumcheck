package indirect

import (
	"fmt"

	"enumbyte"
)

func NonExhaustiveList() {
	var x enumbyte.Letter = 99 // want "implicit conversion of 99 to enumbyte.Letter"
	x = 88                     // want "implicit conversion of 88 to enumbyte.Letter"
	switch x {                 // want "missing cases Delta, Eta and Gamma"
	case enumbyte.Alpha:
		fmt.Println("alpha")
	case enumbyte.Beta, 4: // want "implicit conversion of 4 to enumbyte.Letter"
		fmt.Println("beta")
	default: // want "enumbyte.Letter shouldn't have a default case"
		fmt.Println("default")
	}
}
