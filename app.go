package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/juju/fslock"
)

type App struct {
	persistFilePath string
	resources       map[string]interface{}
}

func (t *App) HasResource(path string) bool {
	_, has := t.resources[path]
	return has
}

func (t *App) SetResource(path string, data interface{}) error {
	t.resources[path] = data
	log.Printf("app: resource [%s] updated\n", path)
	return nil
}

func (t *App) GetResource(path string) (interface{}, error) {
	if !t.HasResource(path) {
		return nil, fmt.Errorf("no such resource: [%s]", path)
	}
	log.Printf("app: resource [%s] queried\n", path)
	return t.resources[path], nil
}

func (t *App) RemoveResource(path string) {
	delete(t.resources, path)
	log.Printf("app: resource [%s] removed\n", path)
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
	if json.NewDecoder(file).Decode(&t.resources); err != nil {
		return err
	}
	log.Printf("app: state restored from %s\n", path)
	return nil
}

func NewApp(persistFilePath string) *App {
	return &App{
		persistFilePath: persistFilePath,
		resources:       make(map[string]interface{}),
	}
}
