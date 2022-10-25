package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/juju/fslock"
)

type Resource = map[string]interface{}

type App struct {
	persistFilePath string
	resources       map[string]Resource
}

func (t *App) HasResource(ns string) bool {
	_, has := t.resources[ns]
	return has
}

func (t *App) SetResource(ns string, data Resource) error {
	t.resources[ns] = data
	log.Printf("app: resource [%s] updated\n", ns)
	return nil
}

func (t *App) MergeResource(ns string, data Resource) error {
	var payload Resource
	if !t.HasResource(ns) {
		payload = Resource{}
	} else {
		payload, _ = t.GetResource(ns)
	}
	return t.SetResource(ns, MergeMaps(payload, data))
}

func (t *App) GetResource(ns string) (Resource, error) {
	if !t.HasResource(ns) {
		return nil, fmt.Errorf("no such resource: [%s]", ns)
	}
	log.Printf("app: resource [%s] queried\n", ns)
	return t.resources[ns], nil
}

func (t *App) RemoveResource(ns string) {
	delete(t.resources, ns)
	log.Printf("app: resource [%s] removed\n", ns)
}

func (t *App) ListNamespaces() []string {
	a := []string{}
	for k, _ := range t.resources {
		a = append(a, k)
	}
	return a
}

func (t *App) Persist() error {
	path := t.persistFilePath
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

func (t *App) Restore() error {
	path := t.persistFilePath
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

func NewApp(persistFilePath string) *App {
	return &App{
		persistFilePath: persistFilePath,
		resources:       map[string]Resource{},
	}
}
