package services

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
)

var DefaultHttpClient = sync.OnceValue(func() *http.Client {
	transport := &http.Transport{}

	proxyURL := os.Getenv("PROXY_URL")
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			panic(fmt.Sprintf("Failed to parse PROXY_URL: %v", err))
		}
		transport.Proxy = http.ProxyURL(proxy)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}


	return &http.Client{Transport: transport}
})
