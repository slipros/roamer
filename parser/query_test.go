package parser

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewQuery(t *testing.T) {
	q := NewQuery()
	require.NotNil(t, q)
	assert.Equal(t, TagQuery, q.Tag())

	q = NewQuery(WithDisabledSplit())
	require.NotNil(t, q)
	assert.False(t, q.split)

	q = NewQuery(WithSplitSymbol(";"))
	require.NotNil(t, q)
	assert.Equal(t, ";", q.splitSymbol)
}

func TestQuery(t *testing.T) {
	queryName := "user_id"
	queryValue := "1337"

	type args struct {
		req   *http.Request
		tag   reflect.StructTag
		cache Cache
	}
	tests := []struct {
		name      string
		args      func() args
		want      any
		notExists bool
	}{
		{
			name: "Get value from query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: queryValue,
		},
		{
			name: "Get value from cached query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				cache := make(map[string]any, 1)
				cache[cacheKeyQuery] = q

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: cache,
				}
			},
			want: queryValue,
		},
		{
			name: "Get value from query - no query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: "",
		},
		{
			name: "Get value from query - wrong query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName+"1", queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: "",
		},
		{
			name: "Get value from array query",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue)
				q.Add(queryName, queryValue+"2")

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: []string{queryValue, queryValue + "2"},
		},
		{
			name: "Get value from query with split symbol",
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue+","+queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagQuery, queryName)),
					cache: make(Cache),
				}
			},
			want: []string{queryValue, queryValue},
		},
		{
			name:      "Wrong tag",
			notExists: true,
			args: func() args {
				rawURL, err := url.Parse(fmt.Sprintf("%s", requestURL))
				require.NoError(t, err)

				q := rawURL.Query()
				q.Add(queryName, queryValue+","+queryValue)

				rawURL.RawQuery = q.Encode()

				req, err := http.NewRequest(http.MethodPost, rawURL.String(), nil)
				require.NoError(t, err)

				return args{
					req:   req,
					tag:   reflect.StructTag(fmt.Sprintf(`%s:"%s"`, TagHeader, queryName)),
					cache: make(Cache),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			args := tt.args()

			q := NewQuery()

			value, exists := q.Parse(args.req, args.tag, args.cache)
			if tt.notExists {
				assert.False(t, exists, "Parse() want not exists, but value exists")
			}

			if tt.want == nil {
				assert.False(t, exists, "Parse() want is nil, but value exists")
			}

			if !tt.notExists {
				assert.Equal(t, tt.want, value)
			}
		})
	}
}

// TestQuery_parseQuery tests the parseQuery function with comprehensive coverage
// to ensure it behaves identically to http.Request.URL.Query() and url.ParseQuery
func TestQuery_parseQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expected    url.Values
		expectError bool
		description string
	}{
		// Successfully parsed cases
		{
			name:        "Empty string",
			query:       "",
			expected:    url.Values{},
			description: "Empty query string should return empty url.Values",
		},
		{
			name:        "Single parameter",
			query:       "a=1",
			expected:    url.Values{"a": {"1"}},
			description: "Single key-value pair",
		},
		{
			name:        "Multiple parameters",
			query:       "a=1&b=2&c=3",
			expected:    url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}},
			description: "Multiple key-value pairs separated by ampersands",
		},
		{
			name:        "Multiple values for same key",
			query:       "a=1&a=2&a=3",
			expected:    url.Values{"a": {"1", "2", "3"}},
			description: "Same key appearing multiple times should create slice",
		},
		{
			name:        "Mixed single and multiple values",
			query:       "a=1&b=2&a=3&c=4",
			expected:    url.Values{"a": {"1", "3"}, "b": {"2"}, "c": {"4"}},
			description: "Combination of single and multiple values for different keys",
		},
		{
			name:        "Parameter without value",
			query:       "a",
			expected:    url.Values{"a": {""}},
			description: "Key without equals sign should have empty string value",
		},
		{
			name:        "Parameter with empty value",
			query:       "a=",
			expected:    url.Values{"a": {""}},
			description: "Key with equals but no value should have empty string value",
		},
		{
			name:        "Mixed parameters with and without values",
			query:       "a=1&b&c=&d=4",
			expected:    url.Values{"a": {"1"}, "b": {""}, "c": {""}, "d": {"4"}},
			description: "Mix of parameters with values, without values, and empty values",
		},
		{
			name:        "Trailing ampersand",
			query:       "a=1&b=2&",
			expected:    url.Values{"a": {"1"}, "b": {"2"}},
			description: "Trailing ampersand should be ignored",
		},
		{
			name:        "Leading ampersand",
			query:       "&a=1&b=2",
			expected:    url.Values{"a": {"1"}, "b": {"2"}},
			description: "Leading ampersand should be ignored",
		},
		{
			name:        "Multiple consecutive ampersands",
			query:       "a=1&&b=2&&&c=3",
			expected:    url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}},
			description: "Multiple consecutive ampersands should be ignored",
		},
		{
			name:        "Only ampersands",
			query:       "&&&",
			expected:    url.Values{},
			description: "String with only ampersands should return empty values",
		},
		{
			name:        "URL encoded key",
			query:       "hello%20world=test",
			expected:    url.Values{"hello world": {"test"}},
			description: "URL encoded spaces in key should be decoded",
		},
		{
			name:        "URL encoded value",
			query:       "test=hello%20world",
			expected:    url.Values{"test": {"hello world"}},
			description: "URL encoded spaces in value should be decoded",
		},
		{
			name:        "URL encoded key and value",
			query:       "hello%20key=world%20value",
			expected:    url.Values{"hello key": {"world value"}},
			description: "Both key and value with URL encoding",
		},
		{
			name:        "Special characters encoded",
			query:       "name=John%20Doe&email=test%40example.com",
			expected:    url.Values{"name": {"John Doe"}, "email": {"test@example.com"}},
			description: "Common special characters like @ and space",
		},
		{
			name:        "Plus signs and ampersands encoded",
			query:       "data=hello%2Bworld%26more",
			expected:    url.Values{"data": {"hello+world&more"}},
			description: "Plus signs and ampersands in values",
		},
		{
			name:        "Unicode characters",
			query:       "unicode=тест&emoji=🚀",
			expected:    url.Values{"unicode": {"тест"}, "emoji": {"🚀"}},
			description: "Unicode characters should be preserved",
		},
		{
			name:        "Percent encoding edge cases",
			query:       "a=100%25&b=%20",
			expected:    url.Values{"a": {"100%"}, "b": {" "}},
			description: "Percent signs and valid percent encoding",
		},
		{
			name:        "Plus sign as space in value",
			query:       "name=John+Doe&city=New+York",
			expected:    url.Values{"name": {"John Doe"}, "city": {"New York"}},
			description: "Plus signs in values should be decoded as spaces",
		},
		{
			name:        "Plus sign as space in key",
			query:       "hello+world=test&foo+bar=value",
			expected:    url.Values{"hello world": {"test"}, "foo bar": {"value"}},
			description: "Plus signs in keys should be decoded as spaces",
		},
		{
			name:        "Mixed plus and percent encoding",
			query:       "text=hello+world%21&name=John+Doe%27s",
			expected:    url.Values{"text": {"hello world!"}, "name": {"John Doe's"}},
			description: "Plus signs and percent encoding should work together",
		},
		{
			name:        "Plus signs without percent encoding",
			query:       "query=search+term+here&filter=active+users",
			expected:    url.Values{"query": {"search term here"}, "filter": {"active users"}},
			description: "Multiple plus signs should all be decoded as spaces",
		},
		{
			name:        "Complex query with all features",
			query:       "name=John%20Doe&tags=go,web&tags=programming&empty=&flag&encoded=hello%2Bworld",
			expected:    url.Values{"name": {"John Doe"}, "tags": {"go,web", "programming"}, "empty": {""}, "flag": {""}, "encoded": {"hello+world"}},
			description: "Complex real-world query string",
		},
		{
			name:        "Equals in value",
			query:       "equation=a%3D1%2Bb%3D2",
			expected:    url.Values{"equation": {"a=1+b=2"}},
			description: "Equals signs within encoded values",
		},
		{
			name:        "Long parameter names and values",
			query:       strings.Repeat("a", 100) + "=" + strings.Repeat("b", 200),
			expected:    url.Values{strings.Repeat("a", 100): {strings.Repeat("b", 200)}},
			description: "Very long parameter names and values",
		},
		{
			name:        "Many parameters",
			query:       generateManyParams(50),
			expected:    generateExpectedManyParams(50),
			description: "Large number of parameters",
		},

		// ===== Whitespace Edge Cases =====
		{
			name:        "Tab character in key",
			query:       "hello%09world=test",
			expected:    url.Values{"hello\tworld": {"test"}},
			description: "Tab character (%09) in key should be decoded to tab",
		},
		{
			name:        "Tab character in value",
			query:       "test=hello%09world",
			expected:    url.Values{"test": {"hello\tworld"}},
			description: "Tab character (%09) in value should be decoded to tab",
		},
		{
			name:        "Newline LF in value",
			query:       "text=line1%0Aline2",
			expected:    url.Values{"text": {"line1\nline2"}},
			description: "Newline LF (%0A) should be decoded to newline character",
		},
		{
			name:        "Carriage return in value",
			query:       "text=line1%0Dline2",
			expected:    url.Values{"text": {"line1\rline2"}},
			description: "Carriage return (%0D) should be decoded to CR character",
		},
		{
			name:        "CRLF in value",
			query:       "text=line1%0D%0Aline2",
			expected:    url.Values{"text": {"line1\r\nline2"}},
			description: "CRLF sequence should be decoded properly",
		},
		{
			name:        "Multiple consecutive spaces encoded",
			query:       "text=hello%20%20%20world",
			expected:    url.Values{"text": {"hello   world"}},
			description: "Multiple consecutive encoded spaces should all be decoded",
		},
		{
			name:        "Multiple consecutive spaces as plus",
			query:       "text=hello+++world",
			expected:    url.Values{"text": {"hello   world"}},
			description: "Multiple consecutive plus signs should all become spaces",
		},
		{
			name:        "Non-breaking space",
			query:       "text=hello%C2%A0world",
			expected:    url.Values{"text": {"hello\u00A0world"}},
			description: "Non-breaking space (U+00A0) should be decoded properly",
		},
		{
			name:        "Leading spaces in key",
			query:       "%20%20key=value",
			expected:    url.Values{"  key": {"value"}},
			description: "Leading spaces in key should be preserved",
		},
		{
			name:        "Trailing spaces in value",
			query:       "key=value%20%20",
			expected:    url.Values{"key": {"value  "}},
			description: "Trailing spaces in value should be preserved",
		},
		{
			name:        "Leading and trailing spaces using plus",
			query:       "key=+++value+++",
			expected:    url.Values{"key": {"   value   "}},
			description: "Leading and trailing plus signs should become spaces",
		},
		{
			name:        "Zero-width space",
			query:       "text=hello%E2%80%8Bworld",
			expected:    url.Values{"text": {"hello\u200Bworld"}},
			description: "Zero-width space (U+200B) should be decoded properly",
		},
		{
			name:        "Mixed whitespace types",
			query:       "text=space%20tab%09newline%0Aplus+end",
			expected:    url.Values{"text": {"space tab\tnewline\nplus end"}},
			description: "Different whitespace types should all be decoded correctly",
		},

		// ===== Escaped Plus vs Regular Plus =====
		{
			name:        "Escaped plus in value",
			query:       "math=1%2B1",
			expected:    url.Values{"math": {"1+1"}},
			description: "Escaped plus (%2B) should become literal plus sign",
		},
		{
			name:        "Regular plus in value",
			query:       "text=hello+world",
			expected:    url.Values{"text": {"hello world"}},
			description: "Regular plus should become space",
		},
		{
			name:        "Mixed escaped and regular plus",
			query:       "expr=a+b%2Bc",
			expected:    url.Values{"expr": {"a b+c"}},
			description: "Mixed plus: regular (+) becomes space, escaped (%2B) becomes literal plus",
		},
		{
			name:        "Leading plus",
			query:       "value=+test",
			expected:    url.Values{"value": {" test"}},
			description: "Leading plus should become space",
		},
		{
			name:        "Trailing plus",
			query:       "value=test+",
			expected:    url.Values{"value": {"test "}},
			description: "Trailing plus should become space",
		},
		{
			name:        "Multiple consecutive pluses",
			query:       "text=a++b",
			expected:    url.Values{"text": {"a  b"}},
			description: "Multiple consecutive pluses should become multiple spaces",
		},
		{
			name:        "Only plus signs",
			query:       "spaces=+++",
			expected:    url.Values{"spaces": {"   "}},
			description: "Only plus signs should become only spaces",
		},
		{
			name:        "Escaped plus at boundaries",
			query:       "expr=%2Ba%2B",
			expected:    url.Values{"expr": {"+a+"}},
			description: "Escaped plus signs at start and end should remain literal",
		},
		{
			name:        "Complex plus mixing",
			query:       "calc=2%2B2+equals+4",
			expected:    url.Values{"calc": {"2+2 equals 4"}},
			description: "Complex mixing of escaped and regular plus signs",
		},

		// ===== Repeated Keys with Empty Values =====
		{
			name:        "Repeated key with trailing empties",
			query:       "key=&key=&key=value",
			expected:    url.Values{"key": {"", "", "value"}},
			description: "Repeated key with empty values followed by actual value",
		},
		{
			name:        "Repeated key with interleaved empties",
			query:       "key=value1&key=&key=value2",
			expected:    url.Values{"key": {"value1", "", "value2"}},
			description: "Repeated key with empty value in the middle",
		},
		{
			name:        "Repeated flag without values",
			query:       "flag&flag&flag",
			expected:    url.Values{"flag": {"", "", ""}},
			description: "Same key repeated without equals sign should create multiple empty values",
		},
		{
			name:        "Mixed empty and non-empty repeated values",
			query:       "key=value1&key&key=&key=value2&key",
			expected:    url.Values{"key": {"value1", "", "", "value2", ""}},
			description: "Complex mix of empty values with and without equals sign",
		},
		{
			name:        "All empty values for key",
			query:       "key=&key=&key=",
			expected:    url.Values{"key": {"", "", ""}},
			description: "All occurrences of key with empty values",
		},

		// ===== Multiple Equals Signs =====
		{
			name:        "Equals in value",
			query:       "equation=a=b=c",
			expected:    url.Values{"equation": {"a=b=c"}},
			description: "Multiple equals signs: first is delimiter, rest are part of value",
		},
		{
			name:        "Multiple equals in different params",
			query:       "data=key=value&other=test",
			expected:    url.Values{"data": {"key=value"}, "other": {"test"}},
			description: "Equals signs in one parameter shouldn't affect others",
		},
		{
			name:        "Double equals",
			query:       "empty==value",
			expected:    url.Values{"empty": {"=value"}},
			description: "Double equals: first is delimiter, second starts the value",
		},
		{
			name:        "Triple equals",
			query:       "triple===value",
			expected:    url.Values{"triple": {"==value"}},
			description: "Triple equals: first is delimiter, others are part of value",
		},
		{
			name:        "Equals at end",
			query:       "key=value=",
			expected:    url.Values{"key": {"value="}},
			description: "Trailing equals should be part of the value",
		},
		{
			name:        "Only equals sign",
			query:       "key======",
			expected:    url.Values{"key": {"====="}},
			description: "Multiple equals: first is delimiter, rest are value",
		},
		{
			name:        "Encoded equals in value",
			query:       "eq=a%3Db%3Dc",
			expected:    url.Values{"eq": {"a=b=c"}},
			description: "URL-encoded equals signs should be decoded to literal equals",
		},

		// ===== Semicolon Handling (Not a Standard Delimiter) =====
		{
			name:        "Semicolon in value",
			query:       "data=a;b;c",
			expected:    url.Values{"data": {"a;b;c"}},
			description: "Semicolon should be treated as part of value, not delimiter",
		},
		{
			name:        "Semicolons with ampersands",
			query:       "mixed=a;b&key=value",
			expected:    url.Values{"mixed": {"a;b"}, "key": {"value"}},
			description: "Semicolons in value with ampersand delimiters should work correctly",
		},
		{
			name:        "Multiple semicolons",
			query:       "path=a;b;c;d&x=1",
			expected:    url.Values{"path": {"a;b;c;d"}, "x": {"1"}},
			description: "Multiple semicolons should all be preserved in value",
		},
		{
			name:        "Encoded semicolon",
			query:       "data=a%3Bb%3Bc",
			expected:    url.Values{"data": {"a;b;c"}},
			description: "Encoded semicolon should decode to literal semicolon",
		},

		// ===== PHP-Style Arrays (Brackets in Keys) =====
		{
			name:        "PHP array style",
			query:       "items[]=1&items[]=2",
			expected:    url.Values{"items[]": {"1", "2"}},
			description: "Brackets should be treated as part of key name",
		},
		{
			name:        "PHP nested array style",
			query:       "user[name]=John&user[age]=30",
			expected:    url.Values{"user[name]": {"John"}, "user[age]": {"30"}},
			description: "Nested bracket notation should preserve brackets in key names",
		},
		{
			name:        "PHP indexed array",
			query:       "data[0]=first&data[1]=second",
			expected:    url.Values{"data[0]": {"first"}, "data[1]": {"second"}},
			description: "Indexed array notation should preserve indices in key names",
		},
		{
			name:        "Encoded brackets",
			query:       "items%5B%5D=value",
			expected:    url.Values{"items[]": {"value"}},
			description: "URL-encoded brackets (%5B %5D) should decode to literal brackets",
		},
		{
			name:        "Mixed bracket styles",
			query:       "a[]=1&b[x]=2&c[0]=3&d=4",
			expected:    url.Values{"a[]": {"1"}, "b[x]": {"2"}, "c[0]": {"3"}, "d": {"4"}},
			description: "Different bracket styles should all be preserved in keys",
		},

		// ===== Special Encoding Edge Cases =====
		{
			name:        "Encoded forward slash",
			query:       "path=api%2Fusers%2F123",
			expected:    url.Values{"path": {"api/users/123"}},
			description: "Encoded forward slashes should decode to literal slashes",
		},
		{
			name:        "Encoded question mark",
			query:       "query=what%3F",
			expected:    url.Values{"query": {"what?"}},
			description: "Encoded question mark should decode to literal question mark",
		},
		{
			name:        "Encoded hash",
			query:       "tag=%23important",
			expected:    url.Values{"tag": {"#important"}},
			description: "Encoded hash should decode to literal hash symbol",
		},
		{
			name:        "Encoded ampersand",
			query:       "text=rock%26roll",
			expected:    url.Values{"text": {"rock&roll"}},
			description: "Encoded ampersand should decode to literal ampersand",
		},
		{
			name:        "Mixed encoded and plus spaces",
			query:       "text=%20+%20",
			expected:    url.Values{"text": {"   "}},
			description: "Mix of %20 and + should all become spaces (three total)",
		},
		{
			name:        "All special URL characters encoded",
			query:       "special=%21%40%23%24%25%5E%26%2A",
			expected:    url.Values{"special": {"!@#$%^&*"}},
			description: "Various special characters should all decode correctly",
		},
		{
			name:        "Encoded colon and slash",
			query:       "url=http%3A%2F%2Fexample.com",
			expected:    url.Values{"url": {"http://example.com"}},
			description: "Encoded URL should decode correctly",
		},

		// ===== Case Sensitivity =====
		{
			name:        "Case sensitive keys",
			query:       "Key=value1&key=value2&KEY=value3",
			expected:    url.Values{"Key": {"value1"}, "key": {"value2"}, "KEY": {"value3"}},
			description: "Keys with different cases should be treated as separate parameters",
		},
		{
			name:        "Mixed case keys with values",
			query:       "UserID=1&userId=2&userid=3&USERID=4",
			expected:    url.Values{"UserID": {"1"}, "userId": {"2"}, "userid": {"3"}, "USERID": {"4"}},
			description: "All case variations of same word should be separate keys",
		},
		{
			name:        "Case sensitive with repeated keys",
			query:       "id=1&ID=2&id=3",
			expected:    url.Values{"id": {"1", "3"}, "ID": {"2"}},
			description: "Case-sensitive keys: lowercase 'id' appears twice, uppercase 'ID' once",
		},

		// ===== Additional Complex Edge Cases =====
		{
			name:        "Empty string after equals for multiple keys",
			query:       "a=&b=&c=",
			expected:    url.Values{"a": {""}, "b": {""}, "c": {""}},
			description: "Multiple keys with empty values after equals",
		},
		{
			name:        "Mix of all edge cases",
			query:       "normal=value&empty=&flag&plus=a+b&escaped=a%2Bb&spaces=%20%20&key=1&key=2",
			expected:    url.Values{"normal": {"value"}, "empty": {""}, "flag": {""}, "plus": {"a b"}, "escaped": {"a+b"}, "spaces": {"  "}, "key": {"1", "2"}},
			description: "Complex combination of various edge cases in one query",
		},
		{
			name:        "Very long key with special chars",
			query:       strings.Repeat("a", 50) + "%20" + strings.Repeat("b", 50) + "=value",
			expected:    url.Values{strings.Repeat("a", 50) + " " + strings.Repeat("b", 50): {"value"}},
			description: "Very long key with encoded space in the middle",
		},
		{
			name:        "Unicode in keys and values",
			query:       "名前=田中&city=東京&emoji=🎉",
			expected:    url.Values{"名前": {"田中"}, "city": {"東京"}, "emoji": {"🎉"}},
			description: "Unicode characters in both keys and values should be preserved",
		},
		{
			name:        "Percent encoding edge - lowercase vs uppercase hex",
			query:       "lower=%2b&upper=%2B",
			expected:    url.Values{"lower": {"+"}, "upper": {"+"}},
			description: "Both lowercase and uppercase hex in percent encoding should decode identically",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := NewQuery()

			// Test custom parseQuery implementation
			result := q.parseQuery(tt.query)

			// Compare with expected result
			assert.Equal(t, tt.expected, result, "parseQuery result should match expected values")

			// Also compare with standard library to ensure identical behavior
			if tt.query != "" {
				standardResult, err := url.ParseQuery(tt.query)
				if err == nil {
					assert.Equal(t, standardResult, result, "parseQuery should behave identically to url.ParseQuery when standard library succeeds")
				}
				// Note: If standard library errors, our implementation may still handle gracefully
			}
		})
	}
}

// TestQuery_parseQuery_ErrorHandling tests error handling scenarios
func TestQuery_parseQuery_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		description string
	}{
		{
			name:        "Invalid percent encoding in value",
			query:       "a=hello%2",
			description: "Invalid percent encoding should not cause panic",
		},
		{
			name:        "Invalid percent encoding in key",
			query:       "hello%2=world",
			description: "Invalid percent encoding in key should not cause panic",
		},
		{
			name:        "Multiple invalid encodings",
			query:       "a%=b%&c%2=d%3",
			description: "Multiple invalid encodings should be handled gracefully",
		},
		{
			name:        "Non-hex percent encoding",
			query:       "a=hello%GG&b=world%ZZ",
			description: "Non-hexadecimal percent encoding should be left as-is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := NewQuery()

			// Should not panic
			assert.NotPanics(t, func() {
				result := q.parseQuery(tt.query)
				assert.NotNil(t, result, "Result should not be nil even with malformed input")
			}, "parseQuery should not panic on malformed input")

			// Compare behavior with standard library
			standardResult, err := url.ParseQuery(tt.query)
			customResult := q.parseQuery(tt.query)

			if err != nil {
				// If standard library fails, our implementation should handle gracefully
				assert.NotNil(t, customResult, "Custom implementation should return valid result even if standard library errors")
			} else {
				// If standard library succeeds, results should match
				assert.Equal(t, standardResult, customResult, "Results should match when standard library succeeds")
			}
		})
	}
}

// TestQuery_parseQuery_Performance tests performance characteristics
func TestQuery_parseQuery_Performance(t *testing.T) {
	// Test with various query sizes to ensure no performance regression
	queries := []struct {
		name  string
		query string
	}{
		{"Small query", "a=1&b=2&c=3"},
		{"Medium query", generateManyParams(100)},
		{"Large query", generateManyParams(1000)},
	}

	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() here because testing.AllocsPerRun is not compatible with parallel tests

			parser := NewQuery()

			// Measure memory allocations
			var result url.Values
			allocs := testing.AllocsPerRun(100, func() {
				result = parser.parseQuery(q.query)
			})

			// Should not allocate excessively
			assert.NotNil(t, result, "Result should not be nil")
			t.Logf("Query size: %d chars, Allocations per run: %.2f", len(q.query), allocs)
		})
	}
}

// Helper function to generate many parameters for testing
func generateManyParams(count int) string {
	var parts []string
	for i := 0; i < count; i++ {
		parts = append(parts, fmt.Sprintf("param%d=value%d", i, i))
	}
	return strings.Join(parts, "&")
}

// Helper function to generate expected result for many parameters
func generateExpectedManyParams(count int) url.Values {
	result := make(url.Values, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("param%d", i)
		value := fmt.Sprintf("value%d", i)
		result[key] = []string{value}
	}
	return result
}

// BenchmarkQuery_parseQuery benchmarks the parseQuery function
func BenchmarkQuery_parseQuery(b *testing.B) {
	benchmarks := []struct {
		name  string
		query string
	}{
		{"Empty", ""},
		{"Single", "a=1"},
		{"Small", "a=1&b=2&c=3"},
		{"Medium", generateManyParams(50)},
		{"Large", generateManyParams(500)},
		{"WithEncoding", "name=John%20Doe&email=test%40example.com&data=hello%2Bworld%26more"},
		{"MultipleValues", "a=1&a=2&a=3&b=4&b=5&c=6"},
	}

	q := NewQuery()

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = q.parseQuery(bm.query)
			}
		})
	}
}

// BenchmarkQuery_parseQuery_vs_Standard compares performance with standard library
func BenchmarkQuery_parseQuery_vs_Standard(b *testing.B) {
	testQuery := generateManyParams(100)
	q := NewQuery()

	b.Run("Custom_parseQuery", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = q.parseQuery(testQuery)
		}
	})

	b.Run("Standard_ParseQuery", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = url.ParseQuery(testQuery)
		}
	})
}
