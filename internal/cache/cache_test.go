package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/cache"
	"github.com/Defacto2/server/internal/helper"
)

func TestCache(t *testing.T) {
	t.Cleanup(func() {
		// Remove cacheDB directory
		path := filepath.Join(helper.TmpDir(), cache.SubDir)
		os.RemoveAll(path)
	})

	const key = "key"

	db := cache.Test
	path, err := db.Path()
	if err != nil {
		t.Errorf("Expected cacheDB path without error: %s", err)
	}
	if path == "" {
		t.Errorf("Expected cacheDB path")
	}

	// Add key-value pair to cache database with default expiry time
	err = db.Write(key, "value", cache.ExpiredAt)
	if err != nil {
		t.Errorf("Expected key-value pair to be added to cache database without error: %s", err)
	}

	// Get value from cache database
	value, err := db.Read(key)
	if err != nil {
		t.Errorf("Expected key 'key' to exist in cache database")
	}
	if value != "value" {
		t.Errorf("Expected value 'value', but got '%s'", value)
	}

	// // Delete key from cache database
	err = db.Delete(key)
	if err != nil {
		t.Errorf("Expected key 'key' to be deleted from cache database without error: %s", err)
	}

	// Check that key was deleted from cache database
	value, err = db.Read(key)
	if err == nil {
		t.Errorf("Expected key 'key' to be deleted from cache database")
	}
	if value != "" {
		t.Errorf("Expected value '', but got '%s'", value)
	}

	// Add key-value pair to cache database with no expiry time
	err = db.WriteNoExpire(key, "value")
	if err != nil {
		t.Errorf("Expected key-value pair to be added to cache database without error: %s", err)
	}

	// Get value from cache database
	value, err = db.Read(key)
	if err != nil {
		t.Errorf("Expected key 'key' to exist in cache database")
	}
	if value != "value" {
		t.Errorf("Expected value 'value', but got '%s'", value)
	}
}
