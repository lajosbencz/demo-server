package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const (
	defPersistFile = "persist.json"
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

	flag.StringVar(&persistFile, "f", defPersistFile, "Persist resource state to this file")
	flag.StringVar(&host, "h", defHost, "Host part of address to listen on")
	flag.IntVar(&port, "p", defPort, "Port part of address to listen on")
	flag.BoolVar(&secure, "s", defSecure, "Enable HTTPS with self-signed certificate")
	flag.Parse()

	app := NewApp(persistFile)
	if err := app.Restore(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	srv, err := NewServer(host, port, secure, router)
	if err != nil {
		panic(err)
	}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Write([]byte("<html><head><title>demo-server</title></head><body><h2>demo-server</h2><p>You need to make a POST request!</p></body></html>"))
			return
		}
		var data AddResource
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			httpError(w, err)
			return
		}
		if app.HasResource(data.Name) && !data.Overwrite {
			httpError(w, fmt.Errorf("resource already exists: [%s]", data.Name))
			return
		}
		err = app.SetResource(data.Name, data.Value)
		if err != nil {
			httpError(w, err)
			return
		}
		res, err := app.GetResource(data.Name)
		if err != nil {
			httpError(w, err)
			return
		}
		httpSuccess(w, map[string]interface{}{
			"error":   false,
			"payload": res,
		})
	})

	router.HandleFunc("/{resourcePath}", func(w http.ResponseWriter, r *http.Request) {
		path := mux.Vars(r)["resourcePath"]
		if r.Method == http.MethodGet {
			if !app.HasResource(path) {
				w.WriteHeader(404)
				return
			}
			data, err := app.GetResource(path)
			if err != nil {
				httpError(w, err)
				return
			}
			httpSuccess(w, map[string]interface{}{
				"error":   false,
				"payload": data,
			})
		} else if r.Method == http.MethodDelete {
			if !app.HasResource(path) {
				httpError(w, fmt.Errorf("no such resource: [%s]", path))
				return
			}
			data, err := app.GetResource(path)
			if err != nil {
				httpError(w, err)
				return
			}
			app.RemoveResource(path)
			httpSuccess(w, map[string]interface{}{
				"error":   false,
				"payload": data,
			})
		}
	})

	srv.Serve()
	proto := "http"
	if secure {
		proto = "https"
	}
	log.Printf("server listening on %s://%s\n", proto, srv.Server.Addr)
	srv.Wait()
	log.Println("shutting server down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		if err := app.Persist(); err != nil {
			log.Println("ERR:", err.Error())
		}
		cancel()
	}()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown failed: %+v", err)
	}
	log.Println("server exited properly")
}
