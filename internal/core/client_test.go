package core_test

import (
	"errors"
	"testing"

	"github.com/ludanortmun/peco/internal/core"
)

type TestStore struct {
	items map[uint]core.Todo
	curr  uint
}

func NewTestStore() *TestStore {
	return &TestStore{
		items: make(map[uint]core.Todo),
		curr:  1,
	}
}

func (store *TestStore) Create(item core.Todo) (core.Todo, error) {
	store.items[store.curr] = item
	item.Id = store.curr
	store.curr++
	return item, nil
}

func (store *TestStore) Update(item core.Todo) (core.Todo, error) {
	_, ok := store.items[item.Id]
	if !ok {
		return core.Todo{}, errors.New(core.ErrNotFound)
	}
	store.items[item.Id] = item
	return item, nil
}

func (store *TestStore) Get(id uint) (core.Todo, bool) {
	item, ok := store.items[id]
	return item, ok
}

func (store *TestStore) List() ([]core.Todo, error) {
	var todos []core.Todo
	for _, item := range store.items {
		todos = append(todos, item)
	}
	return todos, nil
}

func (store *TestStore) Delete(id uint) error {
	_, ok := store.items[id]
	if ok {
		delete(store.items, id)
	}
	return nil
}

func TestGetNotFound(t *testing.T) {
	store := NewTestStore()
	client := core.NewPecoClient(store)

	_, err := client.Get(1)
	if err == nil {
		t.Fatal("Expected error when getting non-existent item, got nil")
	}
	if err.Error() != core.ErrNotFound {
		t.Fatalf("Expected error message '%s', got '%s'", core.ErrNotFound, err.Error())
	}
}

func TestGet(t *testing.T) {
	store := NewTestStore()
	store.items[1] = core.Todo{Id: 1, Content: "Test Todo"}

	client := core.NewPecoClient(store)

	todo, err := client.Get(1)
	if err != nil {
		t.Fatalf("Unexpected error when getting existing item: %v", err)
	}
	if todo.Id != 1 || todo.Content != "Test Todo" {
		t.Fatalf("Unexpected todo item returned: %+v", todo)
	}
}

func TestSaveNewItem(t *testing.T) {
	store := NewTestStore()
	client := core.NewPecoClient(store)

	newTodo := core.Todo{Content: "New Todo"}
	savedTodo, err := client.Save(newTodo)
	if err != nil {
		t.Fatalf("Unexpected error when saving new item: %v", err)
	}
	if savedTodo.Id == 0 {
		t.Fatalf("Expected saved todo to have a non-zero ID, got %d", savedTodo.Id)
	}
	if savedTodo.Content != newTodo.Content {
		t.Fatalf("Expected saved todo content to be '%s', got '%s'", newTodo.Content, savedTodo.Content)
	}
}

func TestSaveNewItemWithNotFoundId(t *testing.T) {
	store := NewTestStore()
	client := core.NewPecoClient(store)

	newTodo := core.Todo{Id: 2, Content: "New Todo"}
	_, err := client.Save(newTodo)
	if err == nil {
		t.Fatal("Expected error when getting non-existent item, got nil")
	}
	if err.Error() != core.ErrNotFound {
		t.Fatalf("Expected error message '%s', got '%s'", core.ErrNotFound, err.Error())
	}
}

func TestSaveExistingItem(t *testing.T) {
	newTodo := core.Todo{Id: 1, Content: "New Todo"}
	store := NewTestStore()
	store.items[1] = newTodo

	client := core.NewPecoClient(store)

	newTodo.Content = "updated"
	savedTodo, err := client.Save(newTodo)

	if err != nil {
		t.Fatalf("Unexpected error when saving new item: %v", err)
	}
	if savedTodo.Id != 1 {
		t.Fatalf("Expected saved todo to have same ID, got %d", savedTodo.Id)
	}
	if savedTodo.Content != "updated" {
		t.Fatal("Expected existing todo content to be updated")
	}
}

func TestSaveNewItemEmptyContent(t *testing.T) {
	newTodo := core.Todo{Content: "New Todo"}
	store := NewTestStore()
	store.items[1] = newTodo

	client := core.NewPecoClient(store)

	newTodo.Content = ""
	_, err := client.Save(newTodo)
	if err == nil {
		t.Fatal("Expected error with invalid input, got nil")
	}
	if err.Error() != core.ErrInvalidInput {
		t.Fatalf("Expected error message '%s', got '%s'", core.ErrNotFound, err.Error())
	}
}

func TestDeleteItem(t *testing.T) {
	store := NewTestStore()
	store.items[1] = core.Todo{Id: 1, Content: "Test Todo"}

	client := core.NewPecoClient(store)

	err := client.Delete(1)
	if err != nil {
		t.Fatalf("Unexpected error when deleting existing item: %v", err)
	}

	_, ok := store.Get(1)
	if ok {
		t.Fatal("Expected item to be deleted, but it still exists")
	}
}

func TestList(t *testing.T) {
	store := NewTestStore()
	store.items[1] = core.Todo{Id: 1, Content: "Test Todo 1"}
	store.items[2] = core.Todo{Id: 2, Content: "Test Todo 2"}

	client := core.NewPecoClient(store)

	todos, err := client.List()
	if err != nil {
		t.Fatalf("Unexpected error when listing items: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("Expected 2 items in the list, got %d", len(todos))
	}
}
