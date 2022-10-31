package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/juju/fslock"
)

// Namespace is an indentifier to which a Resource is assigned
type Namespace = string

// Resource is a generic dictionary stored under a Namespace
type Resource = map[string]interface{}

// App is responsible for storing state in-memory and loading/saving it to persistent storage
type App struct {
	persistFilePath string
	resources       map[Namespace]Resource
}

// HasResource checks if a Namespace already exists
func (t *App) HasResource(ns Namespace) bool {
	_, has := t.resources[ns]
	return has
}

// SetResource creates or overwrites data under a Namespace
func (t *App) SetResource(ns Namespace, data Resource) error {
	t.resources[ns] = data
	log.Printf("app: resource [%s] updated\n", ns)
	return nil
}

// MergeResource recursively merges data under a Namespace
func (t *App) MergeResource(ns Namespace, data Resource) error {
	var payload Resource
	if !t.HasResource(ns) {
		payload = Resource{}
	} else {
		payload, _ = t.GetResource(ns)
	}
	return t.SetResource(ns, MergeMaps(payload, data))
}

// GetResource returns the data under a Namespace
func (t *App) GetResource(ns Namespace) (Resource, error) {
	if !t.HasResource(ns) {
		return nil, fmt.Errorf("no such resource: [%s]", ns)
	}
	log.Printf("app: resource [%s] queried\n", ns)
	return t.resources[ns], nil
}

// RemoveResource deletes the data under a Namespace
func (t *App) RemoveResource(ns Namespace) {
	delete(t.resources, ns)
	log.Printf("app: resource [%s] removed\n", ns)
}

// ListNamespaces lists all the available Namespaces
func (t *App) ListNamespaces() []Namespace {
	a := []Namespace{}
	for k := range t.resources {
		a = append(a, k)
	}
	return a
}

// Persist saves the state of the application to disk
func (t *App) Persist() error {
	path := t.persistFilePath
	if path == "" {
		return nil
	}
	lock := fslock.New(path)
	err := lock.TryLock()
	if err != nil {
		return err
	}
	defer lock.Unlock()
	file, err := os.OpenFile(path, os.O_TRUNC|os.O_WRONLY, 0664)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = json.NewEncoder(file).Encode(t.resources); err != nil {
		return err
	}
	log.Printf("app: state persisted to %s\n", path)
	return nil
}

// Restore restores the state of the application from disk
func (t *App) Restore() error {
	path := t.persistFilePath
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Printf("app: no state to restore from %s\n", path)
		return nil
	}
	lock := fslock.New(path)
	err := lock.TryLock()
	if err != nil {
		return err
	}
	defer lock.Unlock()
	file, err := os.OpenFile(path, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}
	defer file.Close()
	if json.NewDecoder(file).Decode(&t.resources); err != nil {
		return err
	}
	log.Printf("app: state restored from %s\n", path)
	return nil
}

// NewApp creates a new App using the specified file path to persists state to
func NewApp(persistFilePath string) *App {
	return &App{
		persistFilePath: persistFilePath,
		resources:       map[Namespace]Resource{},
	}
}
