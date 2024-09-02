// Package cache provides a lightweight engine for storing key/value pairs.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Defacto2/helper"
	"github.com/rosedblabs/rosedb/v2"
)

type Cache int // Cache is the type of storage engine.

const (
	PouetVote         Cache = iota // data cache for the Pouet website, API requests
	PouetProduction                // data cache for invalid Pouet productions, API requests
	DemozooProduction              // data cache for invalid Demozoo productions, API requests
	Test                           // test cache
)

// String returns the name of the cache.
func (c Cache) String() string {
	return [...]string{
		"pouet",
		"pouetproduction",
		"demozooproduction",
		"test",
	}[c]
}

const (
	ExpiredAt = 7 * 24 * time.Hour // The expiry time for storage engine entries.
	SubDir    = "cacheDB"          // The name of the storage engine subdirectory.
)

// Path returns the absolute path to the storage engine directory.
// If the directory does not exist it will be created.
func (c Cache) Path() (string, error) {
	tmp := filepath.Join(os.TempDir(), SubDir, c.String())
	_, err := os.Stat(tmp)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("cache path %s: %w", tmp, err)
	}
	if err == nil {
		return tmp, nil
	}
	err = os.MkdirAll(tmp, helper.DirWriteReadRead)
	if err != nil {
		return "", fmt.Errorf("cache path, make directory %w", err)
	}
	return tmp, nil
}

// Write writes a key/value pair to the storage engine.
// The key/value pair will be deleted after the ttl time duration has elapsed.
// If ttl is 0 then the key/value pair will immediately expire.
func (c Cache) Write(key, value string, ttl time.Duration) error {
	var err error
	options := rosedb.DefaultOptions
	options.DirPath, err = c.Path()
	if err != nil {
		return fmt.Errorf("cache write %w", err)
	}
	db, err := rosedb.Open(options)
	if err != nil {
		return fmt.Errorf("cache write open rosedb %w", err)
	}
	defer db.Close()

	if err := db.PutWithTTL([]byte(key), []byte(value), ttl); err != nil {
		return fmt.Errorf("cache write save to rosedb %w", err)
	}
	return nil
}

// WriteNoExpire writes a key/value pair to the storage engine.
// The key/value pair will not expire.
func (c Cache) WriteNoExpire(key, value string) error {
	var err error
	options := rosedb.DefaultOptions
	options.DirPath, err = c.Path()
	if err != nil {
		return fmt.Errorf("cache write no expire %w", err)
	}
	db, err := rosedb.Open(options)
	if err != nil {
		return fmt.Errorf("cache write no expire open rosedb %w", err)
	}
	defer db.Close()

	if err := db.Put([]byte(key), []byte(value)); err != nil {
		return fmt.Errorf("cache write no expire save to rosedb %w", err)
	}
	return nil
}

// Read returns value from the storage engine.
func (c Cache) Read(id string) (string, error) {
	path, err := c.Path()
	if err != nil {
		return "", fmt.Errorf("cache read %w", err)
	}
	options := rosedb.DefaultOptions
	options.DirPath = path
	db, err := rosedb.Open(options)
	if err != nil {
		return "", fmt.Errorf("cache read open rosedb %w", err)
	}
	defer db.Close()
	key := []byte(id)
	value, err := db.Get(key)
	if err != nil {
		return "", fmt.Errorf("cache read from rosedb %q: %w", key, err)
	}
	return string(value), nil
}

// Delete deletes a key/value pair from the storage engine.
func (c Cache) Delete(id string) error {
	path, err := c.Path()
	if err != nil {
		return fmt.Errorf("cache delete %w", err)
	}
	options := rosedb.DefaultOptions
	options.DirPath = path
	db, err := rosedb.Open(options)
	if err != nil {
		return fmt.Errorf("cache delete open rosedb %w", err)
	}
	defer db.Close()

	key := []byte(id)
	if err := db.Delete(key); err != nil {
		return fmt.Errorf("cache delete %q: %w", key, err)
	}
	return nil
}
