package alt

import (
	"fmt"

	"github.com/loov/checkenum/testdata"
)

func Example() {
	var x testdata.GreekLetterChecked = 99
	x = 88
	switch x {
	case testdata.Alpha:
		fmt.Println("alpha")
	case testdata.Beta, 4:
		fmt.Println("beta")
	default:
		fmt.Println("default")
	}
}
