package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rosedblabs/rosedb/v2"
)

type Column string // Column Family name.

const (
	Pouet Column = "pouet" // Pouet Column Family.
	Test  Column = "test"  // Test Column Family.
)

const (
	DirMode   = 0755               // Directory permissions.
	ExpiredAt = 7 * 24 * time.Hour // The expiry time for cacheDB entries.
	SubDir    = "cacheDB"          // The name of the cacheDB subdirectory.
)

// Path returns the absolute path to the cacheDB directory.
// If the directory does not exist it will be created.
func (c Column) Path() (string, error) {
	tmp := filepath.Join(os.TempDir(), SubDir, string(c))
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

// Write writes a key/value pair to the cacheDB.
// The key/value pair will be deleted after the ttl time duration has elapsed.
// If ttl is 0 then the key/value pair will immediately expire.
func (c Column) Write(key, value string, ttl time.Duration) error {
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

// WriteNoExpire writes a key/value pair to the cacheDB.
// The key/value pair will not expire.
func (c Column) WriteNoExpire(key, value string) error {
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

// Read returns value from the cacheDB.
func (c Column) Read(id string) (string, error) {
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

// Delete deletes a key/value pair from the cacheDB.
func (c Column) Delete(id string) error {
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
