package calculator

// Operation describes a calculator operation by its arity and implementation.
type Operation struct {
	// Arity is the number of operands the operation uses: 1 or 2.
	Arity int
	// Apply runs the operation. When Arity is 1, b is ignored.
	Apply func(a, b float64) (float64, error)
}

var operations = map[string]Operation{
	"add":        {Arity: 2, Apply: Add},
	"subtract":   {Arity: 2, Apply: Subtract},
	"multiply":   {Arity: 2, Apply: Multiply},
	"divide":     {Arity: 2, Apply: Divide},
	"power":      {Arity: 2, Apply: Power},
	"sqrt":       {Arity: 1, Apply: func(a, _ float64) (float64, error) { return Sqrt(a) }},
	"percentage": {Arity: 2, Apply: Percentage},
}

// Lookup returns the operation registered under name. The boolean is false
// when name is not a known operation.
func Lookup(name string) (Operation, bool) {
	op, ok := operations[name]
	return op, ok
}
