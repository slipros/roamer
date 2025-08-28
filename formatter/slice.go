package formatter

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	rerr "github.com/slipros/roamer/err"
)

const (
	// TagSlice is the struct tag name used for slice formatting.
	TagSlice = "slice"
)

// Slice is a formatter for slice values.
type Slice struct{}

// NewSlice creates a Slice formatter.
func NewSlice() *Slice {
	return &Slice{}
}

// Tag returns the name of the struct tag that this formatter handles.
func (f *Slice) Tag() string {
	return TagSlice
}

// Format applies slice formatters to a field value based on the struct tag.
func (f *Slice) Format(tag reflect.StructTag, ptr any) error {
	tagValue, ok := tag.Lookup(TagSlice)
	if !ok {
		return nil
	}

	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.Wrapf(rerr.NotSupported, "slice formatter for %T", ptr)
	}

	rules := strings.Split(tagValue, SplitSymbol)
	for _, rule := range rules {
		name, arg := parseRule(rule)
		switch name {
		case "unique":
			if err := applyUnique(val.Elem()); err != nil {
				return err
			}
		case "sort":
			if err := applySort(val.Elem(), false); err != nil {
				return err
			}
		case "sort_desc":
			if err := applySort(val.Elem(), true); err != nil {
				return err
			}
		case "compact":
			if err := applyCompact(val.Elem()); err != nil {
				return err
			}
		case "limit":
			if err := applyLimit(val.Elem(), arg); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyUnique(slice reflect.Value) error {
	if slice.Len() == 0 {
		return nil
	}

	seen := make(map[any]struct{})
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
