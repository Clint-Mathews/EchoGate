package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

func TestApiKeyAuthMiddleware(t *testing.T) {
	// 1. Setup: Define a mock token to assert against
	testToken := "super-secret-test-token"

	// Mock your utils package or configuration layer here.
	viper.Set("x_internal_token", testToken)

	// 2. Create a dummy "next" handler to verify the request passes through on success
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Wrap the dummy handler with your middleware
	handlerToTest := ApiKeyAuthMiddleware(nextHandler)

	// 3. Define the test scenarios
	tests := []struct {
		name           string
		headerKey      string
		headerValue    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Authorized - Valid API Key",
			headerKey:      XInternalTokenKey, // This assumes XInternalTokenKey is accessible in the same package
			headerValue:    testToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Unauthorized - Invalid API Key",
			headerKey:      XInternalTokenKey,
			headerValue:    "wrong-token-value",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "401 Unauthorized\n", // http.Error automatically appends a newline
		},
		{
			name:           "Unauthorized - Missing Header entirely",
			headerKey:      "Some-Other-Header",
			headerValue:    testToken,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "401 Unauthorized\n",
		},
		{
			name:           "Unauthorized - Empty Header value",
			headerKey:      XInternalTokenKey,
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "401 Unauthorized\n",
		},
	}

	// 4. Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP request
			req := httptest.NewRequest(http.MethodGet, "http://localhost/any-route", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}

			// Create an HTTP response recorder (implements http.ResponseWriter)
			rr := httptest.NewRecorder()

			// Serve the request through the middleware
			handlerToTest.ServeHTTP(rr, req)

			// Assert Response Code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Assert Response Body
			if rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
