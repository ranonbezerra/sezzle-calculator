// Package calculator implements the arithmetic operations exposed by the API
// as pure functions over float64.
package calculator

import (
	"errors"
	"math"
)

var (
	// ErrDivisionByZero is returned when the divisor is zero.
	ErrDivisionByZero = errors.New("division by zero")
	// ErrNegativeSqrt is returned when taking the square root of a negative number.
	ErrNegativeSqrt = errors.New("square root of a negative number")
	// ErrNonFinite is returned when a result is NaN or ±Inf, for example on overflow.
	ErrNonFinite = errors.New("result is not a finite number")
)

// Add returns a + b.
func Add(a, b float64) (float64, error) {
	return finite(a + b)
}

// Subtract returns a - b.
func Subtract(a, b float64) (float64, error) {
	return finite(a - b)
}

// Multiply returns a * b.
func Multiply(a, b float64) (float64, error) {
	return finite(a * b)
}

// Divide returns a / b. It fails with ErrDivisionByZero when b is zero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return finite(a / b)
}

// Power returns a raised to the power b.
func Power(a, b float64) (float64, error) {
	return finite(math.Pow(a, b))
}

// Sqrt returns the square root of a. It fails with ErrNegativeSqrt when a < 0.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSqrt
	}
	return finite(math.Sqrt(a))
}

// Percentage returns a percent of b, that is a*b/100.
func Percentage(a, b float64) (float64, error) {
	return finite(a * b / 100)
}

func finite(x float64) (float64, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, ErrNonFinite
	}
	if x == 0 {
		return 0, nil // normalizes -0 to 0
	}
	return x, nil
}
