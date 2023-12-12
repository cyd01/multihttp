package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/cyd01/multihttp"
)

var (
	target   = "https://httpbin.org"
	addr     = "127.0.0.1:8080"
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func proxy(w http.ResponseWriter, r *http.Request) {
	targetUrl := r.URL
	targetQuery := targetUrl.RawQuery
	targetUrl, _ = url.Parse(target)
	r.URL.Scheme = targetUrl.Scheme
	r.URL.Host = targetUrl.Host
	r.Header.Set("Host", targetUrl.Host)
	r.URL.Path = singleJoiningSlash(targetUrl.Path, r.URL.Path)
	if targetQuery == "" || r.URL.RawQuery == "" {
		r.URL.RawQuery = targetQuery + r.URL.RawQuery
	} else {
		r.URL.RawQuery = targetQuery + "&" + r.URL.RawQuery
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)
	proxy.ModifyResponse = func(req *http.Response) error {
		if len(req.Header.Get("Location")) > 0 {
			req.Header.Set("Location", strings.ReplaceAll(req.Header.Get("Location"), targetUrl.Host, r.Host))
		}

		return nil
	}
	proxy.ServeHTTP(w, r)
	log.Println(r.Method + " " + r.URL.Host + " => " + r.Host + " " + r.URL.Path)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", proxy)

	log.Println("Starting proxy server towards", target, "at", addr)
	log.Fatal(multihttp.MultiListenAndServe(addr, mux, certFile, keyFile))
}
