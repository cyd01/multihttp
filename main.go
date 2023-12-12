package multihttp

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/http2"
)

type Conn struct {
	net.Conn
	b byte
	e error
	f bool
}

func (c *Conn) Read(b []byte) (int, error) {
	if c.f {
		c.f = false
		b[0] = c.b
		if len(b) > 1 && c.e == nil {
			n, e := c.Conn.Read(b[1:])
			if e != nil {
				c.Conn.Close()
			}
			return n + 1, e
		} else {
			return 1, c.e
		}
	}
	return c.Conn.Read(b)
}

type SplitListener struct {
	net.Listener
	config *tls.Config
}

func (l *SplitListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	b := make([]byte, 1)
	_, err = c.Read(b)
	if err != nil {
		c.Close()
		if err != io.EOF {
			return nil, err
		}
	}
	con := &Conn{
		Conn: c,
		b:    b[0],
		e:    err,
		f:    true,
	}
	if b[0] == 22 {
		//log.Println("HTTPS")
		if l.config != nil {
			return tls.Server(con, l.config), nil
		} else {
			return con, nil
		}
	}
	//log.Println("HTTP")
	return con, nil
}

type Server struct {
	http.Server
	Addr              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	l                 *SplitListener
}

func (s *Server) MultiListenAndServe(handler http.Handler, certFile, keyFile string) error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.New("Can not start listener on " + s.Addr)
	}
	return s.MultiServe(ln, handler, certFile, keyFile)
}

func (s *Server) MultiServe(ln net.Listener, handler http.Handler, certFile, keyFile string) error {
	http2.ConfigureServer(&s.Server, &http2.Server{})
	s.l = &SplitListener{Listener: ln}
	s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			handler.ServeHTTP(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
	if (len(certFile) > 0) && (len(keyFile) > 0) && existfile(certFile) && existfile(keyFile) {
		var err error
		config := &tls.Config{
			MinVersion:             tls.VersionTLS13,
			NextProtos:             []string{"h2", "http/1.1"},
			SessionTicketsDisabled: true,
		}
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return errors.New("Can not load X509 key pair")
		}
		s.l.config = config
	} else {
		s.l.config = nil
	}
	return s.Serve(s.l)
}

func MultiListenAndServe(addr string, handler http.Handler, certFile, keyFile string) error {
	server := &Server{Addr: addr}
	http2.ConfigureServer(&server.Server, &http2.Server{})
	return server.MultiListenAndServe(handler, certFile, keyFile)
}

func MultiServe(ln net.Listener, handler http.Handler, certFile, keyFile string) error {
	server := &Server{}
	http2.ConfigureServer(&server.Server, &http2.Server{})
	return server.MultiServe(ln, handler, certFile, keyFile)
}

func existfile(filename string) bool {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
