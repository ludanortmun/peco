package storage

import (
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/ludanortmun/peco/internal/core"
	_ "modernc.org/sqlite"
)

var _ core.Store = SQLiteStore{}

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
		db.Close()
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
	v := t.UTC().Format("2006-01-02 15:04:05.999999999-07:00")
	return &v
}

func (s SQLiteStore) Close() error {
	return s.db.Close()
}

func (s SQLiteStore) Create(item core.Todo) (core.Todo, error) {
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
	result, err := s.db.Exec(`UPDATE todos SET content = ?, due = ?, completed = ? WHERE id = ?`,
		item.Content, timeToString(item.Due), item.State, item.Id)
	if err != nil {
		return core.Todo{}, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return core.Todo{}, err
	}
	if count == 0 {
		return core.Todo{}, errors.New(core.ErrNotFound)
	}
	return item, nil
}

// scanTodo is shared by single-row and list queries.
func scanTodo(row interface{ Scan(...any) error }) (core.Todo, error) {
	var item core.Todo
	var due sql.NullTime
	err := row.Scan(&item.Id, &item.Content, &item.CreatedAt, &due, &item.State)
	if due.Valid {
		item.Due = &due.Time
	}
	return item, err
}

func (s SQLiteStore) Get(id uint) (core.Todo, bool) {
	item, err := scanTodo(s.db.QueryRow(`SELECT id, content, created_at, due, completed FROM todos WHERE id = ?`, id))
	return item, err == nil
}

func (s SQLiteStore) List(includeDone bool) ([]core.Todo, error) {
	rows, err := s.db.Query(`SELECT id, content, created_at, due, completed FROM todos
  WHERE ? OR completed = 0`, includeDone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]core.Todo, 0)
	for rows.Next() {
		item, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.State != b.State {
			return a.State < b.State
		}
		if a.Due == nil && b.Due != nil {
			return false
		}
		if a.Due != nil && b.Due == nil {
			return true
		}
		if a.Due != nil && !a.Due.Equal(*b.Due) {
			return a.Due.Before(*b.Due)
		}
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.Before(b.CreatedAt)
		}
		return a.Id < b.Id
	})
	return items, nil
}

func (s SQLiteStore) Delete(id uint) error {
	_, err := s.db.Exec(`DELETE FROM todos WHERE id = ?`, id)
	return err
}
