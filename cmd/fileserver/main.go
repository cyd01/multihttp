package main

import (
	"log"
	"net/http"

	"github.com/cyd01/multihttp"
)

var (
	addr     = "127.0.0.1:8080"
	dir      = "."
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(dir)))

	log.Println("Starting file server on", dir, "at", addr)
	log.Fatal(multihttp.MultiListenAndServe(addr, mux, certFile, keyFile))
}
