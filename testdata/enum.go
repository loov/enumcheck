package testdata

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
	var x GreekLetterChecked = 99
	x = 88
	switch x {
	case Alpha:
		fmt.Println("alpha")
	case Beta, 4:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}
