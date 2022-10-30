package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func httpError(w http.ResponseWriter, err error) {
	err2 := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   true,
		"message": err.Error(),
	})
	if err2 != nil {
		log.Fatalln(err2.Error())
	}
	log.Println("ERR:", err.Error())
}

func httpSuccess(w http.ResponseWriter, payload interface{}) {
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		httpError(w, err)
	}
}

// From https://golang.org/src/net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

type Server struct {
	http.Server
	Listener net.Listener
	DoneCh   chan os.Signal
}

func (t *Server) Serve() {
	go func() {
		if err := t.Server.Serve(t.Listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

func (t *Server) Wait() {
	<-t.DoneCh
}

func NewServer(host string, port int, secure bool, handler http.Handler) (*Server, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var tlsCfg *tls.Config = nil
	if secure {
		tlsCfg = &tls.Config{}
		tlsCfg.NextProtos = []string{"http/1.1"}
		tlsCfg.Certificates = make([]tls.Certificate, 1)
		cert, err := GenX509KeyPair()
		if err != nil {
			return nil, err
		}
		tlsCfg.Certificates[0] = cert
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	var listener net.Listener
	if secure {
		listener = tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, tlsCfg)
	} else {
		listener = ln
	}

	return &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: handler,
		},
		Listener: listener,
		DoneCh:   done,
	}, nil
}
