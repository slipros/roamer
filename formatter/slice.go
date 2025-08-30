package formatter

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

// defaultSliceFormatters defines the built-in slice formatting functions.
var defaultSliceFormatters = SliceFormatters{
	"sort":      func(slice reflect.Value, _ string) error { return applySort(slice, false) },
	"sort_desc": func(slice reflect.Value, _ string) error { return applySort(slice, true) },
	"unique":    wrapSliceFunc(applyUnique),
	"compact":   wrapSliceFunc(applyCompact),
	"limit":     applyLimit,
}

const (
	// TagSlice is the struct tag name used for slice formatting.
	TagSlice = "slice"
)

// Slice is a formatter for slice values.
type Slice struct {
	formatters SliceFormatters
}

// NewSlice creates a Slice formatter.
func NewSlice(opts ...SliceOptionsFunc) *Slice {
	s := &Slice{
		formatters: make(SliceFormatters),
	}

	for name, fn := range defaultSliceFormatters {
		s.formatters[name] = fn
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// Tag returns the name of the struct tag that this formatter handles.
func (s *Slice) Tag() string {
	return TagSlice
}

// Format applies slice formatters to a field value based on the struct tag.
func (s *Slice) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagSlice)
	if !ok {
		return nil
	}

	return s.format(tagValue, reflect.ValueOf(ptr))
}

// FormatReflectValue applies slice formatters to a field value based on the struct tag.
func (s *Slice) FormatReflectValue(tag reflect.StructTag, val reflect.Value) error {
	tagValue, ok := tag.Lookup(TagSlice)
	if !ok {
		return nil
	}

	return s.format(tagValue, val)
}

func (s *Slice) format(tagValue string, val reflect.Value) error {
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.Wrapf(rerr.NotSupported, "slice formatter for %s", val.Type().String())
	}

	for _, f := range strings.Split(tagValue, SplitSymbol) {
		name, arg := ParseFormatter(f)

		formatter, ok := s.formatters[name]
		if !ok {
			return errors.WithStack(rerr.FormatterNotFound{Tag: TagSlice, Formatter: name})
		}

		if err := formatter(val.Elem(), arg); err != nil {
			return err
		}
	}

	return nil
}

// wrapSliceFunc wraps a simple slice function to match SliceFormatterFunc signature
func wrapSliceFunc(fn func(slice reflect.Value) error) SliceFormatterFunc {
	return func(slice reflect.Value, _ string) error {
		return fn(slice)
	}
}

func applyUnique(slice reflect.Value) error {
	if slice.Len() == 0 {
		return nil
	}

	seen := make(map[any]struct{}, slice.Len())
	newSlice := reflect.MakeSlice(slice.Type(), 0, slice.Len())

	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i).Interface()
		if _, ok := seen[elem]; !ok {
			seen[elem] = struct{}{}
			newSlice = reflect.Append(newSlice, slice.Index(i))
		}
	}

	slice.Set(newSlice)
	return nil
}

func applySort(slice reflect.Value, desc bool) error {
	switch slice.Type().Elem().Kind() {
	case reflect.Int:
		data := slice.Interface().([]int)
		if desc {
			sort.Sort(sort.Reverse(sort.IntSlice(data)))
		} else {
			sort.Ints(data)
		}
	case reflect.String:
		data := slice.Interface().([]string)
		if desc {
			sort.Sort(sort.Reverse(sort.StringSlice(data)))
		} else {
			sort.Strings(data)
		}
	case reflect.Float64:
		data := slice.Interface().([]float64)
		if desc {
			sort.Sort(sort.Reverse(sort.Float64Slice(data)))
		} else {
			sort.Float64s(data)
		}
	default:
		return errors.Wrapf(rerr.NotSupported, "sort formatter for %s", slice.Type())
	}
	return nil
}

func applyCompact(slice reflect.Value) error {
	if slice.Len() == 0 {
		return nil
	}

	newSlice := reflect.MakeSlice(slice.Type(), 0, slice.Len())

	for i := 0; i < slice.Len(); i++ {
		elem := slice.Index(i)
		if !elem.IsZero() {
			newSlice = reflect.Append(newSlice, elem)
		}
	}

	slice.Set(newSlice)
	return nil
}

func applyLimit(slice reflect.Value, arg string) error {
	limit, err := strconv.Atoi(arg)
	if err != nil {
		return errors.Wrapf(err, "invalid limit value: %s", arg)
	}

	if limit < 0 {
		limit = 0
	}

	if slice.Len() > limit {
		slice.Set(slice.Slice(0, limit))
	}

	return nil
}
