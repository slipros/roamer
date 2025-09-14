package roamer

import (
	"reflect"
)

// Formatter is an interface for components that post-process parsed field values
// based on struct tags. Formatters can be used to transform values after they have
// been parsed from the request (e.g., trimming strings, converting case, etc.).
//
// Implementing a custom formatter allows extending the functionality of the roamer
// package to support additional transformations on parsed values.
//
//go:generate mockery --name=Formatter --outpkg=mockroamer --output=./mockroamer
type Formatter interface {
	// Format transforms a field value based on the provided struct tag.
	// The implementation should check if the tag contains relevant formatting
	// instructions and apply them to the value referenced by the pointer.
	//
	// Parameters:
	//   - tag: The struct tag containing formatting instructions.
	//   - ptr: A pointer to the value to be formatted.
	//
	// Returns:
	//   - error: An error if formatting fails, or nil if successful.
	Format(tag reflect.StructTag, ptr any) error

	// Tag returns the name of the struct tag that this formatter handles.
	// For example, a string formatter might return "string",
	// a number formatter might return "number", etc.
	Tag() string
}

// ReflectValueFormatter is an optional interface that can be implemented by a Formatter.
// If a Formatter implements this interface, the FormatReflectValue method will be called instead of Format.
// This allows the Formatter to work directly with a reflect.Value, which can be more efficient in some cases.
//
//go:generate mockery --name=ReflectValueFormatter --outpkg=mockroamer --output=./mockroamer
type ReflectValueFormatter interface {
	// FormatReflectValue transforms a field value directly using a reflect.Value.
	// This method is called instead of Format when a formatter implements this interface,
	// allowing for more efficient operations by avoiding interface{} conversions.
	//
	// Parameters:
	//   - tag: The struct tag containing formatting instructions.
	//   - val: The reflect.Value to be formatted directly.
	//
	// Returns:
	//   - error: An error if formatting fails, or nil if successful.
	FormatReflectValue(tag reflect.StructTag, val reflect.Value) error
}

// Formatters is a map of registered formatters where keys are the tag names
// returned by the Formatter.Tag() method.
type Formatters map[string]Formatter

// ReflectValueFormatters is a map of registered reflect value formatters where keys are
// the tag names returned by the Formatter.Tag() method.
type ReflectValueFormatters map[string]ReflectValueFormatter
