// Package decoder provides components for decoding HTTP request bodies of various content types.
//
// Decoders implement the Decoder interface and are responsible for parsing request bodies
// based on the Content-Type header. Each decoder handles a specific content type and
// populates Go structures accordingly.
//
// # Built-in Decoders
//
//   - JSON: Handles application/json using jsoniter for performance
//   - XML: Handles application/xml using standard library encoding/xml
//   - FormURL: Handles application/x-www-form-urlencoded for HTML forms
//   - MultipartFormData: Handles multipart/form-data for file uploads and complex forms
//
// # Basic Usage
//
//	// Create decoders
//	jsonDecoder := decoder.NewJSON()
//	formDecoder := decoder.NewFormURL()
//
//	// Use with roamer
//	r := roamer.NewRoamer(
//	    roamer.WithDecoders(jsonDecoder, formDecoder),
//	)
//
// # Custom Decoders
//
// Implement the Decoder interface to support custom content types:
//
//	type MyDecoder struct{}
//
//	func (d *MyDecoder) Decode(r *http.Request, ptr any) error {
//	    // Custom decoding logic
//	    return nil
//	}
//
//	func (d *MyDecoder) ContentType() string {
//	    return "application/my-format"
//	}
//
//	func (d *MyDecoder) Tag() string {
//	    return "myformat"
//	}
//
// # Thread Safety
//
// All built-in decoders are safe for concurrent use and should be reused
// across multiple requests for optimal performance.
package decoder
