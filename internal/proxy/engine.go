package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	utils "github.com/Clint-Mathews/EchoGate/internal/config"
	"github.com/Clint-Mathews/EchoGate/internal/middleware"
)

func ProxyServer() error {
	redirectURL, err := url.Parse("http://localhost:11434")
	if err != nil {
		return err
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.SetURL(redirectURL)
			// Strip your custom gateway authentication header(s)
			r.Out.Header.Del("x-api-key")
		},
		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Set("X-Proxy", "go-reverse-proxy")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}
	http.Handle("/", middleware.ApiKeyAuthMiddleware(proxy))
	fmt.Printf("Running Proxy Server on :%d which routes requests to %s \n", utils.GetRESTPort(), redirectURL.String())
	return http.ListenAndServe(fmt.Sprintf(":%d", utils.GetRESTPort()), nil)
}
