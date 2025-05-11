package value

import (
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetInteger_Successfully(t *testing.T) {
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
	}

	tests := []struct {
		name       string
		setupField func() (reflect.Value, any, any) // returns field, value to set, expected result
	}{
		{
			name: "int to string",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, 42, "42"
			},
		},
		{
			name: "int8 to string",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, int8(127), "127"
			},
		},
		{
			name: "uint to string",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				return field, uint(42), "42"
			},
		},
		{
			name: "int to bool - true for positive",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, 1, true
			},
		},
		{
			name: "int to bool - false for zero",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, 0, false
			},
		},
		{
			name: "int to bool - false for negative",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, -1, false // Negative values should be false
			},
		},
		{
			name: "uint to bool - true for non-zero",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, uint(5), true
			},
		},
		{
			name: "uint to bool - false for zero",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				return field, uint(0), false
			},
		},
		{
			name: "int to int",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				return field, 42, 42
			},
		},
		{
			name: "int to int8 - within range",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, 127, int8(127)
			},
		},
		{
			name: "int to int16 - within range",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int16")
				return field, 32767, int16(32767)
			},
		},
		{
			name: "int to uint - positive value",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint")
				return field, 42, uint(42)
			},
		},
		{
			name: "uint64 to uint8 - within range",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint8")
				return field, uint64(255), uint8(255)
			},
		},
		{
			name: "int to float32",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float32")
				return field, 42, float32(42.0)
			},
		},
		{
			name: "uint to float64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				return field, uint(42), 42.0
			},
		},
		{
			name: "int to complex64",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Complex64")
				return field, 42, complex64(complex(42.0, 0))
			},
		},
		{
			name: "uint to complex128",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Complex128")
				return field, uint(42), complex(42.0, 0)
			},
		},
		{
			name: "int to interface",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Interface")
				return field, 42, int64(42)
			},
		},
		{
			name: "uint to interface",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Interface")
				return field, uint(42), uint64(42)
			},
		},
		{
			name: "int to *int - already initialized",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				initialValue := 0
				s.PtrInt = &initialValue
				field := reflect.ValueOf(s).Elem().FieldByName("PtrInt")
				return field, 42, 42
			},
		},
		{
			name: "int to *int - nil pointer",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrInt")
				return field, 42, 42
			},
		},
		{
			name: "int to *string - nil pointer",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrString")
				return field, 42, "42"
			},
		},
		{
			name: "int to *bool - nil pointer, true for positive",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrBool")
				return field, 42, true
			},
		},
		{
			name: "zero int to *bool - nil pointer, false for zero",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrBool")
				return field, 0, false
			},
		},
		{
			name: "negative int to *bool - nil pointer, false for negative",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrBool")
				return field, -5, false
			},
		},
		{
			name: "max int32 value test",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int32")
				return field, int32(math.MaxInt32), int32(math.MaxInt32)
			},
		},
		{
			name: "max int16 from uint16",
			setupField: func() (reflect.Value, any, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int16")
				return field, uint16(math.MaxInt16), int16(math.MaxInt16)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value, expected := tt.setupField()

			var err error
			switch v := value.(type) {
			case int:
				err = SetInteger(field, v)
			case int8:
				err = SetInteger(field, v)
			case int16:
				err = SetInteger(field, v)
			case int32:
				err = SetInteger(field, v)
			case int64:
				err = SetInteger(field, v)
			case uint:
				err = SetInteger(field, v)
			case uint8:
				err = SetInteger(field, v)
			case uint16:
				err = SetInteger(field, v)
			case uint32:
				err = SetInteger(field, v)
			case uint64:
				err = SetInteger(field, v)
			default:
				t.Fatalf("Unsupported test value type: %T", value)
			}

			require.NoError(t, err)

			// Compare the field value with the expected value
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
			case reflect.Complex64, reflect.Complex128:
				actualValue = field.Complex()
				if field.Kind() == reflect.Complex64 {
					actualValue = complex64(actualValue.(complex128))
				}
			case reflect.Interface:
				actualValue = field.Interface()
			case reflect.Ptr:
				// For pointers, we need to check if they're nil and then get the value
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

			assert.Equal(t, expected, actualValue)
		})
	}
}

func TestSetInteger_Failure(t *testing.T) {
	// Define a struct with fields of various types to test setting values
	type TestStruct struct {
		Int8  int8
		Int16 int16
		Int32 int32
		Uint8 uint8
		Uint  uint
		Map   map[string]string
		Slice []string
	}

	tests := []struct {
		name       string
		setupField func() (reflect.Value, any)   // returns field, value to set
		errorCheck func(t *testing.T, err error) // function to check the error
	}{
		{
			name: "int to int8 - value too large",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, 1000 // 1000 exceeds int8 range
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range")
			},
		},
		{
			name: "int to int8 - value too small",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, -1000 // -1000 exceeds int8 range
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range")
			},
		},
		{
			name: "negative int to uint",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint")
				return field, -1 // Negative value can't be assigned to uint
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "negative value")
			},
		},
		{
			name: "int to unsupported type (map)",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Map")
				return field, 42
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
			},
		},
		{
			name: "int to unsupported type (slice)",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Slice")
				return field, 42
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
			},
		},
		{
			name: "uint64 max to int64",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int8")
				return field, uint64(math.MaxUint64) // Too large for int64
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "overflows")
			},
		},
		{
			name: "uint16 overflow into uint8",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Uint8")
				return field, uint16(256) // 256 is just over uint8 max
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "overflows")
			},
		},
		{
			name: "int32 max + 1 to int16",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int16")
				return field, int32(math.MaxInt16) + 1 // Just over int16 max
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range")
			},
		},
		{
			name: "int32 min - 1 to int16",
			setupField: func() (reflect.Value, any) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int16")
				return field, int32(math.MinInt16) - 1 // Just under int16 min
			},
			errorCheck: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "outside the range")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value := tt.setupField()

			var err error
			switch v := value.(type) {
			case int:
				err = SetInteger(field, v)
			case int8:
				err = SetInteger(field, v)
			case int16:
				err = SetInteger(field, v)
			case int32:
				err = SetInteger(field, v)
			case int64:
				err = SetInteger(field, v)
			case uint:
				err = SetInteger(field, v)
			case uint8:
				err = SetInteger(field, v)
			case uint16:
				err = SetInteger(field, v)
			case uint32:
				err = SetInteger(field, v)
			case uint64:
				err = SetInteger(field, v)
			default:
				t.Fatalf("Unsupported test value type: %T", value)
			}

			tt.errorCheck(t, err)
		})
	}
}

// Benchmark comparing the generic version vs the specialized versions and testing various scenarios
func BenchmarkSetInteger(b *testing.B) {
	// Structure for testing with various types
	type TestStruct struct {
		Int         int
		Int8        int8
		Int16       int16
		Int32       int32
		Int64       int64
		Uint        uint
		Uint8       uint8
		Uint16      uint16
		Uint32      uint32
		Uint64      uint64
		String      string
		Float32     float32
		Float64     float64
		Bool        bool
		Complex     complex128
		PtrInt      *int
		PtrString   *string
		Interface   interface{}
		NilPtrInt   *int
		NilPtrBool  *bool
		NilPtrFloat *float64
	}

	// Test different integer sizes to understand performance impact
	smallInt := 42
	mediumInt := 1000000
	largeInt := math.MaxInt32
	smallUint := uint(42)
	largeUint := uint(math.MaxUint32 / 2)

	// Initialize pointer values for reuse
	initInt := 0
	initStr := ""

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

	// Benchmark different signed integer values to same type
	b.Run("SignedInt", func(b *testing.B) {
		// Small value int to various int types
		runBenchmark("Small-Int-To-Int", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Small-Int-To-Int8", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int8")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Small-Int-To-Int16", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int16")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Small-Int-To-Int32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int32")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Small-Int-To-Int64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int64")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		// Medium value int to int
		runBenchmark("Medium-Int-To-Int", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int")
		}, func(field reflect.Value) {
			_ = SetInteger(field, mediumInt)
		})

		// Large value int to int
		runBenchmark("Large-Int-To-Int", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int")
		}, func(field reflect.Value) {
			_ = SetInteger(field, largeInt)
		})
	})

	// Benchmark different unsigned integer values to same type
	b.Run("UnsignedInt", func(b *testing.B) {
		// Small value uint to various uint types
		runBenchmark("Small-Uint-To-Uint", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})

		runBenchmark("Small-Uint-To-Uint8", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint8")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})

		runBenchmark("Small-Uint-To-Uint16", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint16")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})

		runBenchmark("Small-Uint-To-Uint32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint32")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})

		runBenchmark("Small-Uint-To-Uint64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint64")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})

		// Large value uint to uint
		runBenchmark("Large-Uint-To-Uint", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint")
		}, func(field reflect.Value) {
			_ = SetInteger(field, largeUint)
		})
	})

	// Benchmark integer-to-string conversion (commonly used in web requests)
	b.Run("StringConversions", func(b *testing.B) {
		runBenchmark("Small-Int-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Medium-Int-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetInteger(field, mediumInt)
		})

		runBenchmark("Large-Int-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetInteger(field, largeInt)
		})

		runBenchmark("Small-Uint-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})

		runBenchmark("Large-Uint-To-String", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("String")
		}, func(field reflect.Value) {
			_ = SetInteger(field, largeUint)
		})
	})

	// Benchmark integer-to-float conversion
	b.Run("FloatConversions", func(b *testing.B) {
		runBenchmark("Int-To-Float32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float32")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Int-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Large-Int-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetInteger(field, largeInt)
		})

		runBenchmark("Uint-To-Float64", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Float64")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})
	})

	// Benchmark integer-to-bool conversion
	b.Run("BoolConversions", func(b *testing.B) {
		// Test with zero value
		runBenchmark("Zero-Int-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetInteger(field, 0)
		})

		// Test with positive value
		runBenchmark("Positive-Int-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		// Test with negative value
		runBenchmark("Negative-Int-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetInteger(field, -1)
		})

		// Test with zero unsigned value
		runBenchmark("Zero-Uint-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetInteger(field, uint(0))
		})

		// Test with positive unsigned value
		runBenchmark("Positive-Uint-To-Bool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Bool")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})
	})

	// Benchmark other conversions
	b.Run("OtherConversions", func(b *testing.B) {
		runBenchmark("Int-To-Complex", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Complex")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Int-To-Interface", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Interface")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Uint-To-Interface", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Interface")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallUint)
		})
	})

	// Benchmark pointer handling
	b.Run("PointerHandling", func(b *testing.B) {
		// Initialized pointers
		runBenchmark("Int-To-InitializedPtrInt", func() reflect.Value {
			s := &TestStruct{PtrInt: &initInt}
			return reflect.ValueOf(s).Elem().FieldByName("PtrInt")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Int-To-InitializedPtrString", func() reflect.Value {
			s := &TestStruct{PtrString: &initStr}
			return reflect.ValueOf(s).Elem().FieldByName("PtrString")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		// Nil pointers (requires initialization)
		runBenchmark("Int-To-NilPtrInt", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("NilPtrInt")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Int-To-NilPtrBool", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("NilPtrBool")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})

		runBenchmark("Int-To-NilPtrFloat", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("NilPtrFloat")
		}, func(field reflect.Value) {
			_ = SetInteger(field, smallInt)
		})
	})

	// Compare SetInteger with direct reflect operations using standard type conversions
	b.Run("CompareWithReflect", func(b *testing.B) {
		// First group: Int to Int conversion
		b.Run("Int-To-Int", func(b *testing.B) {
			// Our implementation
			b.Run("SetInteger", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetInteger(field, smallInt)
				}
			})

			// Direct reflect operation
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Int")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetInt(int64(smallInt)) // Native reflect operation
				}
			})
		})

		// Second group: Int to String conversion
		b.Run("Int-To-String", func(b *testing.B) {
			// Our implementation
			b.Run("SetInteger", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetInteger(field, smallInt)
				}
			})

			// Direct string conversion using standard function
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("String")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetString(strconv.FormatInt(int64(smallInt), 10)) // Manual conversion + native reflect
				}
			})
		})

		// Third group: Int to Float64 conversion
		b.Run("Int-To-Float64", func(b *testing.B) {
			// Our implementation
			b.Run("SetInteger", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetInteger(field, smallInt)
				}
			})

			// Direct float conversion
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Float64")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetFloat(float64(smallInt)) // Manual conversion + native reflect
				}
			})
		})

		// Fourth group: Int to Bool conversion
		b.Run("Int-To-Bool", func(b *testing.B) {
			// Our implementation
			b.Run("SetInteger", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetInteger(field, smallInt)
				}
			})

			// Direct bool conversion
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Bool")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.SetBool(smallInt > 0) // Manual conversion + native reflect with the new > 0 logic
				}
			})
		})

		// Fifth group: Int to Interface conversion
		b.Run("Int-To-Interface", func(b *testing.B) {
			// Our implementation
			b.Run("SetInteger", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Interface")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetInteger(field, smallInt)
				}
			})

			// Direct interface assignment
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("Interface")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					field.Set(reflect.ValueOf(int64(smallInt))) // Native reflect
				}
			})
		})

		// Sixth group: Int to Pointer conversion (nil pointer)
		b.Run("Int-To-NilPtr", func(b *testing.B) {
			// Our implementation
			b.Run("SetInteger", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrInt")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					_ = SetInteger(field, smallInt)
				}
			})

			// Manual implementation using native reflect
			b.Run("NativeReflect", func(b *testing.B) {
				s := &TestStruct{}
				field := reflect.ValueOf(s).Elem().FieldByName("NilPtrInt")
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					// Create a new pointer, set its value, and assign to the field
					newPtr := reflect.New(field.Type().Elem())
					newPtr.Elem().SetInt(int64(smallInt))
					field.Set(newPtr)
				}
			})
		})
	})

	// Edge cases and boundary values
	b.Run("EdgeCases", func(b *testing.B) {
		// Max value of different integer types
		runBenchmark("MaxInt8-To-Int16", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int16")
		}, func(field reflect.Value) {
			_ = SetInteger(field, int8(math.MaxInt8))
		})

		runBenchmark("MaxInt16-To-Int32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int32")
		}, func(field reflect.Value) {
			_ = SetInteger(field, int16(math.MaxInt16))
		})

		runBenchmark("MaxUint8-To-Uint16", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint16")
		}, func(field reflect.Value) {
			_ = SetInteger(field, uint8(math.MaxUint8))
		})

		runBenchmark("MaxUint16-To-Uint32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Uint32")
		}, func(field reflect.Value) {
			_ = SetInteger(field, uint16(math.MaxUint16))
		})

		// Min values
		runBenchmark("MinInt8-To-Int16", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int16")
		}, func(field reflect.Value) {
			_ = SetInteger(field, int8(math.MinInt8))
		})

		runBenchmark("MinInt16-To-Int32", func() reflect.Value {
			return reflect.ValueOf(&TestStruct{}).Elem().FieldByName("Int32")
		}, func(field reflect.Value) {
			_ = SetInteger(field, int16(math.MinInt16))
		})
	})
}
