package storage

import (
	"os"
	"path/filepath"
)

// OpenConfig opens the database shared by the CLI and HTTP server.
func OpenConfig() (SQLiteStore, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return SQLiteStore{}, err
	}
	dir := filepath.Join(configDir, "peco")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return SQLiteStore{}, err
	}
	db, err := Open(filepath.Join(dir, "peco.db"))
	if err != nil {
		return SQLiteStore{}, err
	}
	return NewSQLiteStore(db), nil
}
