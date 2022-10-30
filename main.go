package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	defPersistFile = ""
	defHost        = "localhost"
	defPort        = 8080
	defSecure      = false
)

func main() {
	var (
		persistFile string
		host        string
		port        int
		secure      bool
	)

	// parse cli arguments
	flag.StringVar(&persistFile, "f", defPersistFile, "Persist resource state to this file (leave empty to disable)")
	flag.StringVar(&host, "h", defHost, "Host part of address to listen on")
	flag.IntVar(&port, "p", defPort, "Port part of address to listen on")
	flag.BoolVar(&secure, "s", defSecure, "Enable HTTPS with self-signed certificate")
	flag.Parse()

	// create App instance
	app := NewApp(persistFile)
	if err := app.Restore(); err != nil {
		panic(err)
	}

	// create HTTP router
	router := mux.NewRouter()

	// create Server instance
	srv, err := NewServer(host, port, secure, router)
	if err != nil {
		panic(err)
	}

	// setup route for /
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			httpSuccess(w, app.ListNamespaces())
			return
		}
		w.WriteHeader(404)
	})

	// setup catch-all route for namespaced resources
	subRouter := router.PathPrefix("/")
	subRouter.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		var ns Namespace = r.URL.Path[1:]

		// PUT: create/overwrite Resource
		if r.Method == http.MethodPut {
			var payload Resource
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpError(w, err)
				return
			}
			if app.HasResource(ns) {
				httpSuccess(w, Resource{
					"created":   false,
					"namespace": ns,
				})
				return
			}
			if err := app.SetResource(ns, payload); err != nil {
				httpError(w, err)
				return
			}
			httpSuccess(w, Resource{
				"created":   true,
				"namespace": ns,
			})
			return
		}

		// GET: return Resource
		if r.Method == http.MethodGet {
			if !app.HasResource(ns) {
				w.WriteHeader(404)
				return
			}
			payload, _ := app.GetResource(ns)
			if err := json.NewEncoder(w).Encode(payload); err != nil {
				httpError(w, err)
				return
			}
			return
		}

		// POST: update Resource
		if r.Method == http.MethodPost {
			resp := Resource{
				"updated":   false,
				"namespace": ns,
			}
			if !app.HasResource(ns) {
				httpSuccess(w, resp)
				return
			}
			var payload Resource
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				httpError(w, err)
				return
			}
			if err := app.MergeResource(ns, payload); err != nil {
				httpError(w, err)
				return
			}
			resp["updated"] = true
			httpSuccess(w, resp)
			return
		}

		// DELETE: remove Resource
		if r.Method == http.MethodDelete {
			resp := Resource{
				"deleted":   false,
				"namespace": ns,
			}
			if !app.HasResource(ns) {
				httpSuccess(w, resp)
				return
			}
			app.RemoveResource(ns)
			resp["deleted"] = true
			httpSuccess(w, resp)
			return
		}
	})

	// start the Server
	srv.Serve()
	proto := "http"
	if secure {
		proto = "https"
	}
	log.Printf("server listening on %s://%s\n", proto, srv.Server.Addr)
	// wait for the server to start shutting down
	srv.Wait()
	log.Println("shutting server down...")
	// grace period is 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// cleanup tasks
	defer func() {
		if err := app.Persist(); err != nil {
			log.Println("ERR:", err.Error())
		}
		cancel()
		log.Println("cleanup done")
		// fin.
	}()
	// start closing connections
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %+v", err)
	}
	log.Println("server exited properly")
}
