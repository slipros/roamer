package value

import (
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFloat_Successfully(t *testing.T) {
	// Define a struct with fields of various types to test setting values
	type TestStruct struct {
		String     string
		Bool       bool
		Int        int
		Int8       int8
		Int16      int16
		Int32      int32
		Int64      int64
		Uint       uint
		Uint8      uint8
		Uint16     uint16
		Uint32     uint32
		Uint64     uint64
		Float32    float32
		Float64    float64
		Complex64  complex64
		Complex128 complex128
		Interface  any
		// Pointer fields
		PtrString     *string
		PtrInt        *int
		PtrFloat64    *float64
		PtrComplex128 *complex128
		// Initially nil pointer fields to test initialization
		NilPtrInt    *int
		NilPtrString *string
		NilPtrBool   *bool
		NilPtrFloat  *float64
	}

	tests := []struct {
		name       string
		setupField func() (reflect.Value, any, any) // Returns field, value to set, expected result
	}{
		{
			name: "float to string - regular value",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, 42.5, "42.5" // Regular format for normal values
			},
		},
		{
			name: "float to string - NaN",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, math.NaN(), "NaN" // Special handling for NaN
			},
		},
		{
			name: "float to string - positive infinity",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, math.Inf(1), "+Inf" // Special handling for +Inf
			},
		},
		{
			name: "float to string - negative infinity",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, math.Inf(-1), "-Inf" // Special handling for -Inf
			},
		},
		{
			name: "float to string - zero value",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, 0.0, "0" // Zero should be formatted as "0"
			},
		},
		{
			name: "float to string - very large value",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				largeVal := 1.234e10
				return field, largeVal, strconv.FormatFloat(largeVal, 'E', -1, 64) // Scientific for large values
			},
		},
		{
			name: "float to string - very small value",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				smallVal := 1.234e-10
				return field, smallVal, strconv.FormatFloat(smallVal, 'E', -1, 64) // Scientific for small values
			},
		},
		{
			name: "float to bool - positive",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, 42.5, true // Positive float to true
			},
		},
		{
			name: "float to bool - zero",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, 0.0, false // Zero float to false
			},
		},
		{
			name: "float to bool - negative",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, -42.5, false // Negative float to false
			},
		},
		{
			name: "float to int - truncation",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				return field, 42.9, 42 // Should truncate to 42
			},
		},
		{
			name: "float to int8 - within range",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, 127.9, int8(127) // Max int8 value
			},
		},
		{
			name: "negative float to int16",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int16")
				return field, -128.1, int16(-128) // Negative truncation
			},
		},
		{
			name: "float to uint - positive",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint")
				return field, 42.9, uint(42) // Positive truncation
			},
		},
		{
			name: "float to float32 - no overflow",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float32")
				return field, 3.14159, float32(3.14159) // No overflow, but may lose precision
			},
		},
		{
			name: "float to float64 - precision",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				return field, math.Pi, math.Pi // Full precision
			},
		},
		{
			name: "float to complex64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Complex64")
				return field, 42.5, complex64(complex(42.5, 0)) // Real part only
			},
		},
		{
			name: "float to complex128",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Complex128")
				return field, 42.5, complex(42.5, 0) // Real part only
			},
		},
		{
			name: "float to interface",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Interface")
				return field, 42.5, 42.5 // Same value
			},
		},
		{
			name: "float to initialized pointer to float64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				initial := 0.0
				s.PtrFloat64 = &initial
				field := reflect.ValueOf(s).Elem().FieldByName("PtrFloat64")
				return field, 42.5, 42.5 // Override existing value
			},
		},
		{
			name: "float to nil pointer to int",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrInt")
				return field, 42.5, 42 // Should initialize and truncate
			},
		},
		{
			name: "float to nil pointer to string",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrString")
				return field, 42.5, "42.5" // Should initialize and format
			},
		},
		{
			name: "float to nil pointer to bool - positive",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrBool")
				return field, 42.5, true // Should initialize and set true
			},
		},
		{
			name: "float to nil pointer to bool - negative",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrBool")
				return field, -42.5, false // Should initialize and set false
			},
		},
		{
			name: "float to nil pointer to float64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrFloat")
				return field, 42.5, 42.5 // Should initialize with exact value
			},
		},
		{
			name: "NaN to float32",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float32")
				nanValue := float32(math.NaN())
				return field, nanValue, nanValue
			},
		},
		{
			name: "positive infinity to float64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				infValue := math.Inf(1)
				return field, infValue, infValue
			},
		},
		{
			name: "negative infinity to float64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				infValue := math.Inf(-1)
				return field, infValue, infValue
			},
		},
		{
			name: "NaN to complex128",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Complex128")
				nanValue := math.NaN()
				return field, nanValue, complex(nanValue, nanValue)
			},
		},
		{
			name: "tiny positive float to int - truncation to zero",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				return field, 0.1, 0 // Very small numbers are truncated to 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value, expected := tt.setupField()

			var err error
			switch v := value.(type) {
			case float32:
				err = SetFloat(field, v)
			case float64:
				err = SetFloat(field, v)
			default:
				t.Fatalf("Unsupported test value type: %T", value)
			}

			require.NoError(t, err)

			// Verify the field value
			var actualValue any
			switch field.Kind() {
			case reflect.String:
				actualValue = field.String()
			case reflect.Bool:
				actualValue = field.Bool()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				actualValue = field.Int()
				// Convert to expected type for proper comparison
				switch expected.(type) {
				case int:
					actualValue = int(actualValue.(int64))
				case int8:
					actualValue = int8(actualValue.(int64))
				case int16:
					actualValue = int16(actualValue.(int64))
				case int32:
					actualValue = int32(actualValue.(int64))
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				actualValue = field.Uint()
				// Convert to expected type for proper comparison
				switch expected.(type) {
				case uint:
					actualValue = uint(actualValue.(uint64))
				case uint8:
					actualValue = uint8(actualValue.(uint64))
				case uint16:
					actualValue = uint16(actualValue.(uint64))
				case uint32:
					actualValue = uint32(actualValue.(uint64))
				}
			case reflect.Float32, reflect.Float64:
				actualValue = field.Float()
				if field.Kind() == reflect.Float32 {
					actualValue = float32(actualValue.(float64))
				}

				// Special handling for NaN and Inf
				expFloat, expOk := expected.(float64)
				actFloat, actOk := actualValue.(float64)
				if expOk && actOk && (math.IsNaN(expFloat) || math.IsInf(expFloat, 0)) {
					switch {
					case math.IsNaN(expFloat):
						assert.True(t, math.IsNaN(actFloat), "Expected NaN")
					case math.IsInf(expFloat, 1):
						assert.True(t, math.IsInf(actFloat, 1), "Expected +Infinity")
					case math.IsInf(expFloat, -1):
						assert.True(t, math.IsInf(actFloat, -1), "Expected -Infinity")
					}
					return
				}

				// For float32
				expFloat32, expOk32 := expected.(float32)
				actFloat32, actOk32 := actualValue.(float32)
				if expOk32 && actOk32 && (math.IsNaN(float64(expFloat32)) || math.IsInf(float64(expFloat32), 0)) {
					switch {
					case math.IsNaN(float64(expFloat32)):
						assert.True(t, math.IsNaN(float64(actFloat32)), "Expected NaN")
					case math.IsInf(float64(expFloat32), 1):
						assert.True(t, math.IsInf(float64(actFloat32), 1), "Expected +Infinity")
					case math.IsInf(float64(expFloat32), -1):
						assert.True(t, math.IsInf(float64(actFloat32), -1), "Expected -Infinity")
					}
					return
				}

			case reflect.Complex64, reflect.Complex128:
				actualValue = field.Complex()
				if field.Kind() == reflect.Complex64 {
					actualValue = complex64(actualValue.(complex128))
				}

				// Special handling for complex values with NaN components
				expComplex, expOk := expected.(complex128)
				actComplex, actOk := actualValue.(complex128)
				if expOk && actOk {
					expReal := real(expComplex)
					expImag := imag(expComplex)
					actReal := real(actComplex)
					actImag := imag(actComplex)

					// Check if either component is NaN or Inf
					if math.IsNaN(expReal) || math.IsInf(expReal, 0) ||
						math.IsNaN(expImag) || math.IsInf(expImag, 0) {

						if math.IsNaN(expReal) {
							assert.True(t, math.IsNaN(actReal), "Expected real part to be NaN")
						}
						if math.IsNaN(expImag) {
							assert.True(t, math.IsNaN(actImag), "Expected imaginary part to be NaN")
						}
						if math.IsInf(expReal, 1) {
							assert.True(t, math.IsInf(actReal, 1), "Expected real part to be +Infinity")
						}
						if math.IsInf(expReal, -1) {
							assert.True(t, math.IsInf(actReal, -1), "Expected real part to be -Infinity")
						}
						return
					}
				}

				// Similar handling for complex64 if needed

			case reflect.Interface:
				actualValue = field.Interface()
			case reflect.Ptr:
				// For pointers, check and extract the value
				if field.IsNil() {
					t.Fatalf("Expected pointer to be initialized, but it's nil")
				} else {
					elem := field.Elem()
					switch elem.Kind() {
					case reflect.Int:
						actualValue = int(elem.Int())
					case reflect.String:
						actualValue = elem.String()
					case reflect.Float64:
						actualValue = elem.Float()
					case reflect.Complex128:
						actualValue = elem.Complex()
					case reflect.Bool:
						actualValue = elem.Bool()
					default:
						t.Fatalf("Unsupported pointer element type: %s", elem.Kind())
					}
				}
			default:
				t.Fatalf("Unsupported field type for assertion: %s", field.Kind())
			}

			// For floating-point comparisons, use approximate equality
			switch e := expected.(type) {
			case float32, float64:
				assert.InDelta(t, e, actualValue, 1e-7, "Values should be approximately equal")
			default:
				assert.Equal(t, expected, actualValue)
			}
		})
	}
}

func TestSetFloat_Failure(t *testing.T) {
	// Define a struct with fields of various types to test failure cases
	type TestStruct struct {
		Int8      int8
		Int16     int16
		Uint8     uint8
		Uint16    uint16
		Map       map[string]string
		Slice     []string
		Float32   float32
		Interface any
	}

	tests := []struct {
		name       string
		setupField func() (reflect.Value, any)   // Returns field, value to set
		errorCheck func(t *testing.T, err error) // Function to check the error
	}{
		{
			name: "float to int8 - value too large",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, float64(1000) // 1000 exceeds int8 range
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range")
			},
		},
		{
			name: "float to int8 - value too small",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, float64(-1000) // -1000 exceeds int8 range
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range")
			},
		},
		{
			name: "negative float to uint",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint8")
				return field, float64(-1.5) // Negative value can't be assigned to uint
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "negative value")
			},
		},
		{
			name: "float to unsupported type (map)",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Map")
				return field, float64(42.5)
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
			},
		},
		{
			name: "float to unsupported type (slice)",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Slice")
				return field, float64(42.5)
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
			},
		},
		{
			name: "float overflow to float32",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float32")
				return field, float64(math.MaxFloat32) * 2 // Exceeds float32 range
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range of float32")
			},
		},
		{
			name: "float to non-settable field",
			setupField: func() (reflect.Value, any) {
				type privateStruct struct {
					privateField float64
				}
				s := &privateStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("privateField")
				return field, float64(42.5)
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not settable")
			},
		},
		{
			name: "NaN to int",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, math.NaN()
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "NaN")
			},
		},
		{
			name: "Positive Infinity to int",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, math.Inf(1)
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "Infinity")
			},
		},
		{
			name: "Negative Infinity to int",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, math.Inf(-1)
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "Infinity")
			},
		},
		{
			name: "NaN to uint",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint8")
				return field, math.NaN()
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "NaN")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setupField()

			var err error
			switch v := value.(type) {
			case float32:
				err = SetFloat(field, v)
			case float64:
				err = SetFloat(field, v)
			default:
				t.Fatalf("Unsupported test value type: %T", value)
			}

			tt.errorCheck(t, err)
		})
	}
}

// Benchmark for SetFloat performance
func BenchmarkSetFloat(b *testing.B) {
	// Structure for testing with various types
	type TestStruct struct {
		Int         int
		String      string
		Float32     float32
		Float64     float64
		Bool        bool
		Complex     complex128
		PtrInt      *int
		PtrString   *string
		PtrFloat    *float64
		Interface   any
		NilPtrInt   *int
		NilPtrFloat *float64
	}

	// Test different float sizes to understand performance impact
	smallFloat := 42.5
	largeFloat := 1.234e100
	negativeFloat := -42.5

	// Initialize pointer values for reuse
	initInt := 0
	initStr := ""
	initFloat := 0.0

	// Common benchmark helper to reduce boilerplate
	runBenchmark := func(name string, setupField func() reflect.Value, setValue func(reflect.Value)) {
		b.Run(name, func(b *testing.B) {
			field := setupField()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				setValue(field)
			}
		})
	}

	// Benchmark different float values to various types
	b.Run("FloatToNumeric", func(b *testing.B) {
		// Float to int conversions
		runBenchmark("Float-To-Int", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		// Float to float conversions
		runBenchmark("Float-To-Float32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float32")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		runBenchmark("Float-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		// Large and negative values
		runBenchmark("LargeFloat-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetFloat(field, largeFloat)
		})

		runBenchmark("NegativeFloat-To-Int", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int")
		}, func(field reflect.Value) {
			_ = SetFloat(field, negativeFloat)
		})
	})

	// Benchmark string and bool conversions
	b.Run("FloatToOtherTypes", func(b *testing.B) {
		// Float to string conversion
		runBenchmark("Float-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		// Large float to string (scientific notation)
		runBenchmark("LargeFloat-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetFloat(field, largeFloat)
		})

		// Float to bool conversions
		runBenchmark("PositiveFloat-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		runBenchmark("NegativeFloat-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetFloat(field, negativeFloat)
		})

		runBenchmark("ZeroFloat-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetFloat(field, 0.0)
		})
	})

	// Benchmark pointer handling
	b.Run("FloatToPointer", func(b *testing.B) {
		// Initialized pointers
		runBenchmark("Float-To-InitializedPtrInt", func() reflect.Value {
			s := &TestStruct{PtrInt: &initInt}
			return reflect.ValueOf(s).Elem().FieldByName("PtrInt")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		runBenchmark("Float-To-InitializedPtrString", func() reflect.Value {
			s := &TestStruct{PtrString: &initStr}
			return reflect.ValueOf(s).Elem().FieldByName("PtrString")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		runBenchmark("Float-To-InitializedPtrFloat", func() reflect.Value {
			s := &TestStruct{PtrFloat: &initFloat}
			return reflect.ValueOf(s).Elem().FieldByName("PtrFloat")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		// Nil pointers (requires initialization)
		runBenchmark("Float-To-NilPtrInt", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("NilPtrInt")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})

		runBenchmark("Float-To-NilPtrFloat", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("NilPtrFloat")
		}, func(field reflect.Value) {
			_ = SetFloat(field, smallFloat)
		})
	})

	// Compare SetFloat with direct reflect operations
	b.Run("CompareWithNativeReflect", func(b *testing.B) {
		// Float to Float64 comparison
		b.Run("Float-To-Float64", func(b *testing.B) {
			// Our implementation
			b.Run("SetFloat", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetFloat(field, smallFloat)
				}
			})

			// Direct reflect operation
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetFloat(smallFloat) // Native reflect operation
				}
			})
		})

		// Float to String comparison
		b.Run("Float-To-String", func(b *testing.B) {
			// Our implementation
			b.Run("SetFloat", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetFloat(field, smallFloat)
				}
			})

			// Direct string conversion using standard function
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetString(strconv.FormatFloat(smallFloat, 'f', -1, 64)) // Manual conversion + native reflect
				}
			})
		})

		// Float to Int comparison
		b.Run("Float-To-Int", func(b *testing.B) {
			// Our implementation
			b.Run("SetFloat", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetFloat(field, smallFloat)
				}
			})

			// Direct int conversion
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetInt(int64(smallFloat)) // Manual conversion + native reflect
				}
			})
		})

		// Float to Bool comparison
		b.Run("Float-To-Bool", func(b *testing.B) {
			// Our implementation
			b.Run("SetFloat", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetFloat(field, smallFloat)
				}
			})

			// Direct bool conversion
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetBool(smallFloat > 0) // Manual conversion + native reflect
				}
			})
		})
	})

	// Benchmark special float values (NaN, Infinity)
	b.Run("SpecialFloatValues", func(b *testing.B) {
		nanValue := math.NaN()
		posInfValue := math.Inf(1)
		negInfValue := math.Inf(-1)
		tinyValue := 1e-10 // Will be truncated to 0 for int types

		// NaN to different types
		runBenchmark("NaN-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetFloat(field, nanValue)
		})

		runBenchmark("NaN-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetFloat(field, nanValue)
		})

		// +Infinity to different types
		runBenchmark("PosInf-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetFloat(field, posInfValue)
		})

		runBenchmark("PosInf-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetFloat(field, posInfValue)
		})

		// -Infinity to different types
		runBenchmark("NegInf-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetFloat(field, negInfValue)
		})

		runBenchmark("NegInf-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetFloat(field, negInfValue)
		})

		// Very small value that will be truncated
		runBenchmark("TinyFloat-To-Int", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int")
		}, func(field reflect.Value) {
			_ = SetFloat(field, tinyValue)
		})
	})
}
