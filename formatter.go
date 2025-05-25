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

// Formatters is a map of registered formatters where keys are the tag names
// returned by the Formatter.Tag() method.
type Formatters map[string]Formatter
