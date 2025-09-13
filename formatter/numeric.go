package formatter

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// defaultNumericFormatters defines the built-in numeric formatting functions.
var defaultNumericFormatters = NumericFormatters{
	"abs":   wrapNumericFunc(applyAbs),
	"round": wrapNumericFunc(applyRound),
	"ceil":  wrapNumericFunc(applyCeil),
	"floor": wrapNumericFunc(applyFloor),
	"min":   applyMin,
	"max":   applyMax,
}

const (
	// TagNumeric is the struct tag name used for numeric formatting.
	TagNumeric = "numeric"
)

// Numeric is a formatter for numeric values.
// It applies transformations to numeric fields based on the "numeric" struct tag.
type Numeric struct {
	formatters NumericFormatters
}

// NewNumeric creates a Numeric formatter.
func NewNumeric(opts ...NumericOptionsFunc) *Numeric {
	n := &Numeric{
		formatters: make(NumericFormatters),
	}

	for name, fn := range defaultNumericFormatters {
		n.formatters[name] = fn
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// Tag returns the name of the struct tag that this formatter handles.
func (n *Numeric) Tag() string {
	return TagNumeric
}

// Format applies numeric formatters to a field value based on the struct tag.
func (n *Numeric) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagNumeric)
	if !ok {
		return nil
	}

	for _, f := range strings.Split(tagValue, SplitSymbol) {
		name, arg := ParseFormatter(f)

		formatter, ok := n.formatters[name]
		if !ok {
			return errors.WithStack(rerr.FormatterNotFoundError{Tag: TagNumeric, Formatter: name})
		}

		if err := formatter(ptr, arg); err != nil {
			return err
		}
	}

	return nil
}

// wrapNumericFunc wraps a simple numeric function to match NumericFormatterFunc signature.
// This utility function adapts functions that only need the pointer
// to the interface that expects both pointer and argument parameters.
func wrapNumericFunc(fn func(ptr any) error) NumericFormatterFunc {
	return func(ptr any, _ string) error {
		return fn(ptr)
	}
}

// applyMin ensures a numeric value is not less than a specified minimum.
func applyMin(ptr any, arg string) error {
	return applyMinMax(ptr, arg, true)
}

// applyMax ensures a numeric value is not greater than a specified maximum.
func applyMax(ptr any, arg string) error {
	return applyMinMax(ptr, arg, false)
}

// applyMinMax applies either minimum or maximum constraint to a numeric value.
// If isMin is true, applies minimum constraint; otherwise applies maximum constraint.
func applyMinMax(ptr any, arg string, isMin bool) error {
	opName := "max"
	if isMin {
		opName = "min"
	}

	switch v := ptr.(type) {
	case *int:
		m, err := strconv.Atoi(arg)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < m) || (!isMin && *v > m) {
			*v = m
		}
	case *int8:
		m, err := strconv.ParseInt(arg, 10, 8)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < int8(m)) || (!isMin && *v > int8(m)) {
			*v = int8(m)
		}
	case *int16:
		m, err := strconv.ParseInt(arg, 10, 16)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < int16(m)) || (!isMin && *v > int16(m)) {
			*v = int16(m)
		}
	case *int32:
		m, err := strconv.ParseInt(arg, 10, 32)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < int32(m)) || (!isMin && *v > int32(m)) {
			*v = int32(m)
		}
	case *int64:
		m, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < m) || (!isMin && *v > m) {
			*v = m
		}
	case *float32:
		m, err := strconv.ParseFloat(arg, 32)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < float32(m)) || (!isMin && *v > float32(m)) {
			*v = float32(m)
		}
	case *float64:
		m, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < m) || (!isMin && *v > m) {
			*v = m
		}
	default:
		return errors.Wrapf(rerr.NotSupported, "%s formatter for %T", opName, ptr)
	}

	return nil
}

// applyAbs converts a numeric value to its absolute value.
func applyAbs(ptr any) error {
	switch v := ptr.(type) {
	case *int:
		if *v < 0 {
			*v = -*v
		}
	case *int8:
		if *v < 0 {
			*v = -*v
		}
	case *int16:
		if *v < 0 {
			*v = -*v
		}
	case *int32:
		if *v < 0 {
			*v = -*v
		}
	case *int64:
		if *v < 0 {
			*v = -*v
		}
	case *float32:
		*v = float32(math.Abs(float64(*v)))
	case *float64:
		*v = math.Abs(*v)
	default:
		return errors.Wrapf(rerr.NotSupported, "abs formatter for %T", ptr)
	}

	return nil
}

// applyRound rounds a floating-point value to the nearest integer.
func applyRound(ptr any) error {
	return applyFloatFunc(ptr, "round", math.Round)
}

// applyCeil rounds a floating-point value up to the nearest integer.
func applyCeil(ptr any) error {
	return applyFloatFunc(ptr, "ceil", math.Ceil)
}

// applyFloor rounds a floating-point value down to the nearest integer.
func applyFloor(ptr any) error {
	return applyFloatFunc(ptr, "floor", math.Floor)
}

// applyFloatFunc applies a mathematical function to floating-point values.
// The opName parameter is used for error messages to identify the operation.
func applyFloatFunc(ptr any, opName string, fn func(float64) float64) error {
	switch v := ptr.(type) {
	case *float32:
		*v = float32(fn(float64(*v)))
	case *float64:
		*v = fn(*v)
	default:
		return errors.Wrapf(rerr.NotSupported, "%s formatter for %T", opName, ptr)
	}
	return nil
}
