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
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagNumeric, Formatter: name})
		}

		if err := formatter(ptr, arg); err != nil {
			return err
		}
	}

	return nil
}

// wrapNumericFunc wraps a simple numeric function to match NumericFormatterFunc signature
func wrapNumericFunc(fn func(ptr any) error) NumericFormatterFunc {
	return func(ptr any, _ string) error {
		return fn(ptr)
	}
}

func applyMin(ptr any, arg string) error {
	return applyMinMax(ptr, arg, true)
}

func applyMax(ptr any, arg string) error {
	return applyMinMax(ptr, arg, false)
}

func applyMinMax(ptr any, arg string, isMin bool) error {
	opName := "max"
	if isMin {
		opName = "min"
	}

	switch v := ptr.(type) {
	case *int:
		m, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "invalid %s value: %s", opName, arg)
		}
		if (isMin && *v < int(m)) || (!isMin && *v > int(m)) {
			*v = int(m)
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

func applyRound(ptr any) error {
	return applyFloatFunc(ptr, "round", math.Round)
}

func applyCeil(ptr any) error {
	return applyFloatFunc(ptr, "ceil", math.Ceil)
}

func applyFloor(ptr any) error {
	return applyFloatFunc(ptr, "floor", math.Floor)
}

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
