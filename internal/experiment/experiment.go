// Package experiment contains experimental features that may change or be removed in future versions.
// This package is intended for internal use only and should not be imported directly by users.
package experiment

// Experiment is an interface for components that can enable experimental features.
// This interface allows the roamer package to configure experimental optimizations
// in compatible components without knowing their concrete types.
//
// Currently, the only experimental feature is the fast struct field parser,
// which uses optimized reflection techniques to access struct fields more efficiently.
//
// Warning: Implementations of this interface may change between minor versions,
// as the experimental features evolve or are removed.
type Experiment interface {
	// EnableExperimentalFastStructFieldParser enables the use of an experimental
	// fast struct field parser in the implementing component. This can improve
	// performance but may not be as stable as the standard parser.
	//
	// This method is called by the roamer package when the WithExperimentalFastStructFieldParser
	// option is used.
	EnableExperimentalFastStructFieldParser()
}
