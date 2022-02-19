// want package:"enumtype.Expr = {Add | Mul | Div | Value}"
package enumtype

// Expr is an enumerated type.
//
//enumcheck:exhaustive
type Expr interface{}

var _ Expr = Add{}
var _ Expr = Mul{}
var (
	_, _ Expr = Div{}, Value(0)
)

type Add []Expr
type Mul []Expr
type Div []Expr
type Misc struct{}
type Value float64

var invalid = Value(-1)

func Eval(x Expr) (Value, error) {
	switch x := x.(type) { // want "missing cases enumtype.Div"
	case Add:
		total, err := Eval(x[0])
		if err != nil {
			return invalid, err
		}
		for _, v := range x[1:] {
			a, err := Eval(v)
			if err != nil {
				return invalid, err
			}
			total += a
		}
		return total, nil
	case Mul:
		total, err := Eval(x[0])
		if err != nil {
			return invalid, err
		}
		for _, v := range x[1:] {
			a, err := Eval(v)
			if err != nil {
				return invalid, err
			}
			total *= a
		}
		return total, nil
	case Value:
		return x, nil
	case interface{ Name() string }: // want "implicit conversion of interface[{]Name[(][)] string[}] to enumtype.Expr"
		return 0, nil
	default:
		return invalid, nil
	}
}

func Name(x Expr) string {
	switch x.(type) { // want "missing cases enumtype.Div"
	case Add:
		return "Add"
	case Mul:
		return "Mul"
	case Value:
		return "Value"
	case Misc: // want "implicit conversion of enumtype.Misc to enumtype.Expr"
		return "Misc"
	default:
		return "unknown"
	}
}

func ImplicitConversion() {
	var _ Expr = Misc{} // want "implicit conversion of enumtype.Misc to enumtype.Expr"
	_, _ = Eval(Misc{}) // want "implicit conversion of enumtype.Misc to enumtype.Expr"
	_ = Name(Misc{})    // want "implicit conversion of enumtype.Misc to enumtype.Expr"
	Name(Misc{})        // want "implicit conversion of enumtype.Misc to enumtype.Expr"
}

type Struct struct {
	Value Expr
}

func AssignmentToStruct() {
	var s Struct

	s.Value = Misc{} // want "implicit conversion of enumtype.Misc to enumtype.Expr"
	s.Value = Add{}
}

func ExpandedAssignment() {
	var x Expr
	var s Struct

	s.Value, x = Values()
	_, _ = s.Value, x
}

func Values() (a, b Expr) {
	return Add{}, Misc{} // want "implicit conversion of enumtype.Misc to enumtype.Expr"
}

func ValuesX() (a, b Expr) {
	return Values()
}

func Chan() {
	ch := make(chan Expr, 10)
	ch <- Add{}
	ch <- Misc{} // want "implicit conversion of enumtype.Misc to enumtype.Expr"
}

func NamedChan() {
	type ExprChan chan Expr
	ch := make(ExprChan)
	ch <- Add{}
	ch <- Misc{} // want "implicit conversion of enumtype.Misc to enumtype.Expr"
}

func ChanFunc() {
	fn := func() chan Expr { return make(chan Expr, 10) }
	fn() <- Add{}
	fn() <- Misc{} // want "implicit conversion of enumtype.Misc to enumtype.Expr"
}
