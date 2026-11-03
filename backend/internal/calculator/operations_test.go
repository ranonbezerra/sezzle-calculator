package calculator

import (
	"errors"
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		name      string
		a, b      float64
		wantArity int
		want      float64
		wantErr   error
	}{
		{name: "add", a: 6, b: 3, wantArity: 2, want: 9},
		{name: "subtract", a: 6, b: 3, wantArity: 2, want: 3},
		{name: "multiply", a: 6, b: 3, wantArity: 2, want: 18},
		{name: "divide", a: 6, b: 3, wantArity: 2, want: 2},
		{name: "power", a: 6, b: 3, wantArity: 2, want: 216},
		{name: "sqrt", a: 9, b: 123, wantArity: 1, want: 3},
		{name: "percentage", a: 50, b: 30, wantArity: 2, want: 15},
		{name: "divide", a: 1, b: 0, wantArity: 2, wantErr: ErrDivisionByZero},
		{name: "sqrt", a: -1, wantArity: 1, wantErr: ErrNegativeSqrt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, ok := Lookup(tt.name)
			if !ok {
				t.Fatalf("Lookup(%q) not found", tt.name)
			}
			if op.Arity != tt.wantArity {
				t.Errorf("Arity = %d, want %d", op.Arity, tt.wantArity)
			}
			got, err := op.Apply(tt.a, tt.b)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Apply(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestLookupUnknown(t *testing.T) {
	for _, name := range []string{"", "modulo", "Add", " add", "sqrt/"} {
		t.Run(name, func(t *testing.T) {
			if _, ok := Lookup(name); ok {
				t.Errorf("Lookup(%q) found, want not found", name)
			}
		})
	}
}
