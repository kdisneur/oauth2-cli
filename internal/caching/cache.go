package caching

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// Read loads the cache file as JSON and unmarshal it to dest
// It returns an error if it can't unmarshal the cache but not if the cache is not present
func Read(folder string, filename string, dest interface{}) error {
	filepath := path.Join(folder, filename)
	file, err := os.Open(filepath)
	if err != nil {
		// safely ignore if we miss a cache hit
		return nil
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(dest); err != nil {
		return fmt.Errorf("can't decode '%s' cache: %v", filepath, err)
	}

	return nil
}

// Save persists the cache file as JSON from src
// It returns an error if it can't marshal the cache but not if the cache can't be persisted
func Save(folder string, filename string, src interface{}) error {
	filepath := path.Join(folder, filename)
	os.MkdirAll(folder, 0755)

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		// safely ignore if we miss a cache hit
		return nil
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(src); err != nil {
		return fmt.Errorf("can't encode '%s' cache: %v", filepath, err)
	}

	return nil
}
