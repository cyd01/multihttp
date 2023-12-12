package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/cyd01/multihttp"
)

var (
	addr     = "127.0.0.1:8080"
	certFile = "cert.pem"
	keyFile  = "key.pem"
)

func hello(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.URL.Path)
	fmt.Fprintln(w, "Hello world!")
}

func main2() {
	var err error

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)

	server := &multihttp.Server{}
	log.Fatal(server.MultiServe(ln, mux, certFile, keyFile))
}

func main3() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)

	server := &multihttp.Server{
		Addr:              addr,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	log.Fatal(server.MultiListenAndServe(mux, certFile, keyFile))
}

func main4() {
	var err error

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)

	log.Fatal(multihttp.MultiServe(ln, mux, certFile, keyFile))
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)

	log.Println("Starting hello server at", addr)
	log.Fatal(multihttp.MultiListenAndServe(addr, mux, certFile, keyFile))
}
