package proxy

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Clint-Mathews/EchoGate/internal/middleware"
	"github.com/spf13/viper"
)

func TestProxyServer_Integration(t *testing.T) {
	// 1. Isolate the global HTTP Multiplexer (Mux)
	// This prevents registration panics if this test runs multiple times or alongside other tests.
	oldMux := http.DefaultServeMux
	defer func() { http.DefaultServeMux = oldMux }()
	http.DefaultServeMux = http.NewServeMux()

	// 2. Start a mock backend on the hardcoded port 11434 (e.g., mimicking Ollama)
	// We use net.Listen to claim port 11434. If it's already in use (e.g., Ollama is running),
	// we skip the test gracefully with an informative message.
	backendAddr := "127.0.0.1:11430"
	backendListener, err := net.Listen("tcp", backendAddr)
	if err != nil {
		t.Skipf("Skipping test: Port 11434 is already in use (%v). Stop Ollama or any local service on 11434 to run this integration test.", err)
	}
	defer func() {
		if err := backendListener.Close(); err != nil {
			fmt.Println("Error", err)
		}
	}()

	// A channel to capture the request that the proxy forwards to our mock backend
	backendRequestChan := make(chan *http.Request, 1)

	backendServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture the incoming request for assertions
			backendRequestChan <- r
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("backend success"))
		}),
	}
	go func() {
		_ = backendServer.Serve(backendListener)
	}()

	defer func() {
		if err := backendServer.Close(); err != nil {
			fmt.Println("Error", err)
		}
	}()

	// 3. Dynamically allocate a free port for the Proxy Server
	// This avoids hardcoding port 8080 or conflicting with local development servers.
	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to allocate a dynamic port for proxy: %v", err)
	}
	proxyPort := proxyListener.Addr().(*net.TCPAddr).Port
	if err := proxyListener.Close(); err != nil {
		fmt.Println("Error", err)
	} // Close it immediately so ProxyServer can bind to it

	// 4. Set up configuration/environment variables for the tests
	testToken := "super-secret-gateway-token"

	// Set these variables using both environment variables and Viper
	// to make sure your custom utils configuration layer registers them properly.
	redirectURL := "http://" + backendAddr // "http://127.0.0.1:11430"
	t.Setenv("REST_PORT", fmt.Sprintf("%d", proxyPort))
	t.Setenv("X_INTERNAL_TOKEN", testToken) // Adjust this key to match what your middleware expects
	t.Setenv("REDIRECT_URL", redirectURL)
	viper.Set("REST_PORT", proxyPort)
	viper.Set("X_INTERNAL_TOKEN", testToken) // Adjust this key to match what your middleware expects
	viper.Set("REDIRECT_URL", redirectURL)

	// 5. Start the Proxy Server in a background goroutine
	go func() {
		if err := ProxyServer(); err != nil && err != http.ErrServerClosed {
			t.Logf("ProxyServer exited with error: %v", err)
		}
	}()

	// Give the proxy server a brief moment to start up and bind to the port
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{Timeout: 2 * time.Second}
	proxyAddr := fmt.Sprintf("http://127.0.0.1:%d", proxyPort)

	// --- TEST CASE 1: Authorized Request ---
	t.Run("Authorized Request - Proxy Forwards and Strips Token", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, proxyAddr, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		// Send the custom API key to satisfy the middleware
		req.Header.Set(middleware.XInternalTokenKey, testToken)

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request to proxy: %v", err)
		}

		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Println("Error", err)
			}
		}()

		// Verify proxy response
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
		}

		// Verify ModifyResponse injected the custom proxy header
		if gotHeader := resp.Header.Get("X-Proxy"); gotHeader != "go-reverse-proxy" {
			t.Errorf("Expected X-Proxy header 'go-reverse-proxy', got '%s'", gotHeader)
		}

		// Verify backend assertions
		select {
		case backendReq := <-backendRequestChan:
			// Ensure the token was successfully stripped before hitting the backend
			if token := backendReq.Header.Get(middleware.XInternalTokenKey); token != "" {
				t.Errorf("Expected 'x-api-key' to be stripped from the request, but found: %s", token)
			}
			// Verify path was successfully preserved
			if backendReq.URL.Path != "/" {
				t.Errorf("Expected forwarded path '/', got '%s'", backendReq.URL.Path)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout: The proxy did not forward the request to the mock backend")
		}
	})

	// --- TEST CASE 2: Unauthorized Request ---
	t.Run("Unauthorized Request - Middleware Blocks Execution", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, proxyAddr, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		// Send an incorrect token
		req.Header.Set(middleware.XInternalTokenKey, "invalid-token-value")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request to proxy: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Println("Error", err)
			}
		}()

		// Verify middleware blocked the request
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 Unauthorized, got %d", resp.StatusCode)
		}

		// Ensure the backend was never reached
		select {
		case <-backendRequestChan:
			t.Error("Security Breach: Request reached backend even with invalid credentials!")
		case <-time.After(100 * time.Millisecond):
			// Success: Timeout here indicates the request was correctly blocked
		}
	})
}
