package core

import "time"

type State int64

const (
	Pending State = iota
	Done
)

type Todo struct {
	Id        uint       `json:"id"` // non-zero value indicates an existing item
	Content   string     `json:"content"`
	State     State      `json:"state"`
	CreatedAt time.Time  `json:"createdAt"`
	Due       *time.Time `json:"due"`
}

func NewTodo(content string) Todo {
	return Todo{
		Content: content,
		State:   Pending,
	}
}

func NewTodoWithDueDate(content string, due time.Time) Todo {
	due = due.UTC()
	return Todo{
		Content: content,
		State:   Pending,
		Due:     &due,
	}
}
