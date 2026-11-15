package core

import "time"

type State int64

const (
	Pending State = iota
	Done
)

type Todo struct {
	Id        uint // non-zero value indicates an existing item
	Content   string
	State     State
	CreatedAt time.Time
	Due       *time.Time
}

func NewTodo(content string) Todo {
	return Todo{
		Content: content,
		State:   Pending,
	}
}

func NewTodoWithDueDate(content string, due time.Time) Todo {
	return Todo{
		Content: content,
		State:   Pending,
		Due:     &due,
	}
}
