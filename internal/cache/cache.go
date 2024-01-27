// Package cache provides a lightweight engine for storing key/value pairs.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rosedblabs/rosedb/v2"
)

type Cache int

const (
	Pouet Cache = iota // pouet data cache
	Test               // test cache
)

// String returns the name of the cache.
func (c Cache) String() string {
	return [...]string{"pouet", "test"}[c]
}

const (
	DirMode   = 0o755              // Directory permissions.
	ExpiredAt = 7 * 24 * time.Hour // The expiry time for storage engine entries.
	SubDir    = "cacheDB"          // The name of the storage engine subdirectory.
)

// Path returns the absolute path to the storage engine directory.
// If the directory does not exist it will be created.
func (c Cache) Path() (string, error) {
	tmp := filepath.Join(os.TempDir(), SubDir, c.String())
	_, err := os.Stat(tmp)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("%s: %w", tmp, err)
	}
	if err == nil {
		return tmp, nil
	}
	err = os.MkdirAll(tmp, DirMode)
	if err != nil {
		return "", err
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
		return err
	}
	db, err := rosedb.Open(options)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.PutWithTTL([]byte(key), []byte(value), ttl)
}

// WriteNoExpire writes a key/value pair to the storage engine.
// The key/value pair will not expire.
func (c Cache) WriteNoExpire(key, value string) error {
	var err error

	options := rosedb.DefaultOptions
	options.DirPath, err = c.Path()
	if err != nil {
		return err
	}
	db, err := rosedb.Open(options)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Put([]byte(key), []byte(value))
}

// Read returns value from the storage engine.
func (c Cache) Read(id string) (string, error) {
	path, err := c.Path()
	if err != nil {
		return "", err
	}

	options := rosedb.DefaultOptions
	options.DirPath = path
	db, err := rosedb.Open(options)
	if err != nil {
		return "", err
	}
	defer db.Close()

	key := []byte(id)
	value, err := db.Get(key)
	if err != nil {
		return "", fmt.Errorf("%q: %w", key, err)
	}
	return string(value), nil
}

// Delete deletes a key/value pair from the storage engine.
func (c Cache) Delete(id string) error {
	path, err := c.Path()
	if err != nil {
		return err
	}

	options := rosedb.DefaultOptions
	options.DirPath = path
	db, err := rosedb.Open(options)
	if err != nil {
		return err
	}
	defer db.Close()

	key := []byte(id)
	return db.Delete(key)
}
