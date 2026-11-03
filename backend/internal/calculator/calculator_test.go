package calculator

import (
	"errors"
	"math"
	"testing"
)

type binaryCase struct {
	name    string
	a, b    float64
	want    float64
	wantErr error
}

func runBinary(t *testing.T, fn func(a, b float64) (float64, error), tests []binaryCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fn(tt.a, tt.b)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func assertResult(t *testing.T, got float64, err error, want float64, wantErr error) {
	t.Helper()
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAdd(t *testing.T) {
	runBinary(t, Add, []binaryCase{
		{name: "positives", a: 2, b: 3, want: 5},
		{name: "zeros", a: 0, b: 0, want: 0},
		{name: "negatives", a: -2, b: -3, want: -5},
		{name: "mixed signs", a: -2, b: 5, want: 3},
		{name: "decimals", a: 1.5, b: 2.25, want: 3.75},
		{name: "very large values", a: 0x1p1000, b: 0x1p1000, want: 0x1p1001},
		{name: "large opposites cancel", a: math.MaxFloat64, b: -math.MaxFloat64, want: 0},
		{name: "overflow", a: math.MaxFloat64, b: math.MaxFloat64, wantErr: ErrNonFinite},
	})
}

// TestAddReturnsRawFloat64 documents that the calculator does not round:
// 0.1 + 0.2 yields the raw IEEE 754 result. Formatting is the frontend's job.
func TestAddReturnsRawFloat64(t *testing.T) {
	got, err := Add(0.1, 0.2)
	assertResult(t, got, err, 0.30000000000000004, nil)
}

func TestSubtract(t *testing.T) {
	runBinary(t, Subtract, []binaryCase{
		{name: "positives", a: 5, b: 3, want: 2},
		{name: "zeros", a: 0, b: 0, want: 0},
		{name: "negative result", a: 3, b: 5, want: -2},
		{name: "negatives", a: -2, b: -3, want: 1},
		{name: "decimals", a: 5.75, b: 2.5, want: 3.25},
		{name: "very large values", a: 0x1p1001, b: 0x1p1000, want: 0x1p1000},
		{name: "overflow", a: -math.MaxFloat64, b: math.MaxFloat64, wantErr: ErrNonFinite},
	})
}

func TestMultiply(t *testing.T) {
	runBinary(t, Multiply, []binaryCase{
		{name: "positives", a: 4, b: 3, want: 12},
		{name: "by zero", a: 7, b: 0, want: 0},
		{name: "negatives", a: -4, b: -3, want: 12},
		{name: "mixed signs", a: -4, b: 3, want: -12},
		{name: "decimals", a: 1.5, b: 0.5, want: 0.75},
		{name: "very large values", a: 0x1p500, b: 0x1p500, want: 0x1p1000},
		{name: "overflow", a: 1e200, b: 1e200, wantErr: ErrNonFinite},
		{name: "negative overflow", a: -1e200, b: 1e200, wantErr: ErrNonFinite},
	})
}

func TestDivide(t *testing.T) {
	runBinary(t, Divide, []binaryCase{
		{name: "positives", a: 10, b: 4, want: 2.5},
		{name: "zero dividend", a: 0, b: 5, want: 0},
		{name: "negatives", a: -9, b: -3, want: 3},
		{name: "mixed signs", a: -9, b: 3, want: -3},
		{name: "decimals", a: 1, b: 3, want: 1.0 / 3},
		{name: "very large values", a: 0x1p1000, b: 0x1p500, want: 0x1p500},
		{name: "by zero", a: 5, b: 0, wantErr: ErrDivisionByZero},
		{name: "by negative zero", a: 5, b: math.Copysign(0, -1), wantErr: ErrDivisionByZero},
		{name: "zero by zero", a: 0, b: 0, wantErr: ErrDivisionByZero},
		{name: "overflow", a: math.MaxFloat64, b: 0.5, wantErr: ErrNonFinite},
	})
}

func TestPower(t *testing.T) {
	runBinary(t, Power, []binaryCase{
		{name: "positives", a: 2, b: 10, want: 1024},
		{name: "zero exponent", a: 5, b: 0, want: 1},
		{name: "zero to the zero", a: 0, b: 0, want: 1},
		{name: "zero base", a: 0, b: 3, want: 0},
		{name: "negative base odd exponent", a: -2, b: 3, want: -8},
		{name: "negative base even exponent", a: -2, b: 2, want: 4},
		{name: "negative exponent", a: 2, b: -2, want: 0.25},
		{name: "fractional exponent", a: 4, b: 0.5, want: 2},
		{name: "decimal base", a: 1.5, b: 2, want: 2.25},
		{name: "very large values", a: 2, b: 1000, want: 0x1p1000},
		{name: "overflow", a: 10, b: 400, wantErr: ErrNonFinite},
		{name: "negative base fractional exponent", a: -8, b: 1.0 / 3, wantErr: ErrNonFinite},
		{name: "zero to negative exponent", a: 0, b: -1, wantErr: ErrNonFinite},
	})
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr error
	}{
		{name: "perfect square", a: 9, want: 3},
		{name: "zero", a: 0, want: 0},
		{name: "decimal", a: 2.25, want: 1.5},
		{name: "very large value", a: 0x1p1000, want: 0x1p500},
		{name: "negative", a: -4, wantErr: ErrNegativeSqrt},
		{name: "NaN input", a: math.NaN(), wantErr: ErrNonFinite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sqrt(tt.a)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func TestPercentage(t *testing.T) {
	runBinary(t, Percentage, []binaryCase{
		{name: "50 percent of 200", a: 50, b: 200, want: 100},
		{name: "10 percent of 30", a: 10, b: 30, want: 3},
		{name: "over 100 percent", a: 150, b: 40, want: 60},
		{name: "zero percent", a: 0, b: 50, want: 0},
		{name: "percent of zero", a: 50, b: 0, want: 0},
		{name: "negative percent", a: -20, b: 50, want: -10},
		{name: "percent of negative", a: 20, b: -50, want: -10},
		{name: "decimals", a: 12.5, b: 80, want: 10},
		{name: "very large values", a: 50, b: 0x1p1000, want: 0x1p999},
		{name: "overflow", a: 1e300, b: 1e300, wantErr: ErrNonFinite},
	})
}

func TestNegativeZeroIsNormalized(t *testing.T) {
	got, err := Multiply(-1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Signbit(got) {
		t.Errorf("Multiply(-1, 0) = -0, want +0")
	}
}
