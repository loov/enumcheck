package use

import (
	"fmt"

	"def"
)

func Example() {
	var x def.GreekLetterChecked = 99 // want "basic literal declaration to checked enum"
	x = 88                            // want "basic literal assignment to checked enum"
	switch x {                        // want "switch clause missing for Delta, Eta and Gamma"
	case def.Alpha:
		fmt.Println("alpha")
	case def.Beta, 4: // want "basic literal clause for checked enum"
		fmt.Println("beta")
	default: // want "default literal clause for checked enum"
		fmt.Println("default")
	}
}
