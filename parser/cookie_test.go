package parser

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCookie(t *testing.T) {
	h := NewCookie()
	require.NotNil(t, h)
	require.Equal(t, TagCookie, h.Tag())
}

func TestCookie(t *testing.T) {
	cookie := "ref"
	cookieValue := "test"

	type args struct {
		req *http.Request
		tag reflect.StructTag
	}
	tests := []struct {
		name string
		args func() args
		want any
	}{
		{
			name: "Get cookie value from request cookie",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.AddCookie(&http.Cookie{
					Name:  cookie,
					Value: cookieValue,
				})

				return args{
					req: req,
					tag: reflect.StructTag(fmt.Sprintf(`%s:"%s""`, TagCookie, cookie)),
				}
			},
			want: &http.Cookie{
				Name:  cookie,
				Value: cookieValue,
			},
		},
		{
			name: "Get cookie value from request cookie - empty struct tag",
			args: func() args {
				req, err := http.NewRequest(http.MethodPost, requestURL, nil)
				require.NoError(t, err)

				req.AddCookie(&http.Cookie{
					Name:  cookie,
					Value: cookieValue,
				})

				return args{
					req: req,
					tag: "",
				}
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args()

			h := NewCookie()
			value, exists := h.Parse(args.req, args.tag, nil)

			if tt.want == nil && exists {
				t.Errorf("Parse() want is nil, but value exists")
			}

			require.Equal(t, tt.want, value)
		})
	}
}

func TestCookieAssignExtensions(t *testing.T) {
	tests := []struct {
		name string
		// Test setup and verification
		testFunc func(t *testing.T)
	}{
		{
			name: "Successfully - Returns exactly one extension function",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1, "Should return exactly one extension function")
			},
		},
		{
			name: "Successfully - Extension accepts *http.Cookie and assigns cookie.Value to string field",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1)

				testCases := []struct {
					name        string
					cookieValue string
					expectValue string
				}{
					{
						name:        "Simple cookie value",
						cookieValue: "test_value",
						expectValue: "test_value",
					},
					{
						name:        "Empty cookie value",
						cookieValue: "",
						expectValue: "",
					},
					{
						name:        "Cookie value with spaces",
						cookieValue: "  spaced value  ",
						expectValue: "  spaced value  ",
					},
					{
						name:        "Cookie value with special characters",
						cookieValue: "value!@#$%^&*()_+-={}|[]\\:;\"'<>?,./",
						expectValue: "value!@#$%^&*()_+-={}|[]\\:;\"'<>?,./",
					},
					{
						name:        "Unicode cookie value",
						cookieValue: "тест_значение_🍪",
						expectValue: "тест_значение_🍪",
					},
					{
						name:        "JSON-like cookie value",
						cookieValue: `{"key":"value","number":123}`,
						expectValue: `{"key":"value","number":123}`,
					},
					{
						name:        "URL-encoded cookie value",
						cookieValue: "user%40example.com",
						expectValue: "user%40example.com",
					},
				}

				for _, tc := range testCases {
					t.Run(tc.name, func(t *testing.T) {
						cookie := &http.Cookie{
							Name:  "test_cookie",
							Value: tc.cookieValue,
						}

						// Test extension function with cookie
						assignFunc, ok := extensions[0](cookie)
						require.True(t, ok, "Extension should accept *http.Cookie")
						require.NotNil(t, assignFunc, "Should return assignment function")

						// Test assignment to string field
						var result string
						target := reflect.ValueOf(&result).Elem()

						err := assignFunc(target)
						require.NoError(t, err, "Assignment should succeed")
						require.Equal(t, tc.expectValue, result, "Should assign cookie.Value to target field")
					})
				}
			},
		},
		{
			name: "Successfully - Extension assigns cookie.Value to different target types",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1)

				cookie := &http.Cookie{
					Name:  "test_cookie",
					Value: "123",
				}

				assignFunc, ok := extensions[0](cookie)
				require.True(t, ok)
				require.NotNil(t, assignFunc)

				testCases := []struct {
					name        string
					setupTarget func() reflect.Value
					verify      func(t *testing.T, target reflect.Value)
				}{
					{
						name: "String target",
						setupTarget: func() reflect.Value {
							var str string
							return reflect.ValueOf(&str).Elem()
						},
						verify: func(t *testing.T, target reflect.Value) {
							require.Equal(t, "123", target.String())
						},
					},
					{
						name: "Interface{} target",
						setupTarget: func() reflect.Value {
							var iface interface{}
							return reflect.ValueOf(&iface).Elem()
						},
						verify: func(t *testing.T, target reflect.Value) {
							require.Equal(t, "123", target.Interface())
						},
					},
				}

				for _, tc := range testCases {
					t.Run(tc.name, func(t *testing.T) {
						target := tc.setupTarget()
						err := assignFunc(target)
						require.NoError(t, err, "Assignment should succeed")
						tc.verify(t, target)
					})
				}
			},
		},
		{
			name: "Failure - Extension rejects non-cookie values",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1)

				testCases := []struct {
					name  string
					value interface{}
				}{
					{
						name:  "nil value",
						value: nil,
					},
					{
						name:  "string value",
						value: "test_string",
					},
					{
						name:  "int value",
						value: 123,
					},
					{
						name:  "bool value",
						value: true,
					},
					{
						name:  "slice value",
						value: []string{"a", "b", "c"},
					},
					{
						name:  "map value",
						value: map[string]string{"key": "value"},
					},
					{
						name:  "struct value",
						value: struct{ Name string }{Name: "test"},
					},
					{
						name:  "pointer to string",
						value: func() *string { s := "test"; return &s }(),
					},
					{
						name:  "http.Request",
						value: &http.Request{},
					},
					{
						name:  "http.Cookie value (not pointer)",
						value: http.Cookie{Name: "test", Value: "value"},
					},
				}

				for _, tc := range testCases {
					t.Run(tc.name, func(t *testing.T) {
						assignFunc, ok := extensions[0](tc.value)
						require.False(t, ok, "Extension should reject non-cookie value: %T", tc.value)
						require.Nil(t, assignFunc, "Should return nil assignment function for non-cookie value")
					})
				}
			},
		},
		{
			name: "Failure - Assignment error handling",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1)

				cookie := &http.Cookie{
					Name:  "test_cookie",
					Value: "not_a_number",
				}

				assignFunc, ok := extensions[0](cookie)
				require.True(t, ok)
				require.NotNil(t, assignFunc)

				// Test assignment to incompatible types that might cause errors
				testCases := []struct {
					name        string
					setupTarget func() reflect.Value
					expectError bool
				}{
					{
						name: "Assignment to int with non-numeric string",
						setupTarget: func() reflect.Value {
							var num int
							return reflect.ValueOf(&num).Elem()
						},
						expectError: true,
					},
					{
						name: "Assignment to bool with non-boolean string",
						setupTarget: func() reflect.Value {
							var b bool
							return reflect.ValueOf(&b).Elem()
						},
						expectError: true,
					},
				}

				for _, tc := range testCases {
					t.Run(tc.name, func(t *testing.T) {
						target := tc.setupTarget()
						err := assignFunc(target)

						if tc.expectError {
							require.Error(t, err, "Assignment should fail for incompatible type")
						} else {
							require.NoError(t, err, "Assignment should succeed")
						}
					})
				}
			},
		},
		{
			name: "Successfully - Extension works with cookie containing all http.Cookie fields",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1)

				// Create cookie with all possible fields set
				cookie := &http.Cookie{
					Name:     "session_id",
					Value:    "abc123xyz",
					Path:     "/api",
					Domain:   "example.com",
					MaxAge:   3600,
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
				}

				assignFunc, ok := extensions[0](cookie)
				require.True(t, ok, "Extension should accept cookie with all fields")
				require.NotNil(t, assignFunc, "Should return assignment function")

				var result string
				target := reflect.ValueOf(&result).Elem()

				err := assignFunc(target)
				require.NoError(t, err, "Assignment should succeed")
				require.Equal(t, "abc123xyz", result, "Should extract only the Value field, ignoring other cookie attributes")
			},
		},
		{
			name: "Failure - Assignment to invalid reflect.Value causes panic/error",
			testFunc: func(t *testing.T) {
				c := NewCookie()
				extensions := c.AssignExtensions()
				require.Len(t, extensions, 1)

				cookie := &http.Cookie{
					Name:  "test_cookie",
					Value: "test_value",
				}

				assignFunc, ok := extensions[0](cookie)
				require.True(t, ok)
				require.NotNil(t, assignFunc)

				// Test assignment to zero Value should cause panic/error
				// We expect this to panic since assign.String checks target.Type() on zero Value
				defer func() {
					r := recover()
					require.NotNil(t, r, "Assignment to zero Value should panic")
					require.Contains(t, fmt.Sprintf("%v", r), "zero Value", "Panic should mention zero Value")
				}()

				// This should panic
				zeroValue := reflect.Value{}
				_ = assignFunc(zeroValue)

				// If we reach here, the test failed
				t.Error("Expected panic when assigning to zero Value, but none occurred")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
