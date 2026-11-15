package storage

import (
	"database/sql"
	"time"

	"github.com/ludanortmun/peco/internal/core"
	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Create the todos table if it doesn't exist
	_, err = db.Exec(
		`CREATE TABLE IF NOT EXISTS todos(
			id INTEGER,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			due DATETIME,
			completed BOOLEAN NOT NULL DEFAULT 0,
			PRIMARY KEY(id AUTOINCREMENT)
		);`,
	)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewSQLiteStore(db *sql.DB) SQLiteStore {
	return SQLiteStore{db: db}
}

func timeToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	v := t.Format("2006-01-02 15:04:05")
	return &v
}

func (s SQLiteStore) Close() error {
	return s.db.Close()
}

func (s SQLiteStore) Create(item core.Todo) (core.Todo, error) {
	// Implementation for creating a new core.Todo item in the SQLite database
	res, err := s.db.Exec(
		`INSERT INTO todos (content, created_at, due, completed)
		 VALUES (?, ?, ?, ?);`,
		item.Content, timeToString(&item.CreatedAt), timeToString(item.Due), item.State,
	)
	if err != nil {
		return core.Todo{}, err
	}

	id, err := res.LastInsertId()
	item.Id = uint(id)
	return item, err
}

func (s SQLiteStore) Update(item core.Todo) (core.Todo, error) {
	// Implementation for updating an existing core.Todo item in the SQLite database
	return core.Todo{}, nil
}

func (s SQLiteStore) Get(id uint) (core.Todo, bool) {
	// Implementation for retrieving a core.Todo item by ID from the SQLite database
	return core.Todo{}, false
}

func (s SQLiteStore) List() ([]core.Todo, error) {
	// Implementation for listing all core.Todo items from the SQLite database
	return []core.Todo{}, nil
}

func (s SQLiteStore) Delete(id uint) error {
	// Implementation for deleting a core.Todo item by ID from the SQLite database
	return nil
}
