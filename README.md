# enumcheck

***This is still a WIP, so exact behavior may change.***

Analyzer for exhaustive enum switches.

This package reports errors for:

``` go
type Letter byte // enumcheck

const (
	Alpha Letter = iota
	Beta
	Gamma
)

func Switch(x Letter) {
	switch x { // error: "missing cases Beta and Gamma"
	case Alpha:
		fmt.Println("alpha")
	case 4: // error: "implicit conversion of 4 to Letter"
		fmt.Println("beta")
	default: // error: "Letter shouldn't have a default case"
		fmt.Println("default")
	}
}

func Assignment() {
    var x Letter
    x = 123 // error: "implicit conversion of 123 to Letter
}

```

This can also be used with types:

``` go
type Expr interface{} // enumcheck

var _ Expr = Add{}
var _ Expr = Mul{}

type Add []Expr
type Mul []Expr

type Invalid []Expr

func Switch(x Expr) {
	switch x.(type) { // error: "missing cases Mul"
	case Add:
		fmt.Println("alpha")
	case Invalid: // error: "implicit conversion of Invalid to Expr"
		fmt.Println("beta")
	default:
		fmt.Println("unknown")
	}
}

func Assignment() {
	var x Expr
	x = 3 // error: "implicit conversion of 3 to Expr
	_ = x
}
```