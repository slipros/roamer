package chi

/*
func createHTTPServer(t *testing.T, h http.Handler) *url.URL {
	t.Helper()

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	return u
}

func TestNewPath(t *testing.T) {
	tests := []struct {
		name      string
		newRouter func() *chi.Mux
		path      string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest()

			got := NewPath(tt.newRouter())

			if got := NewPath(tt.newRouter()); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/
