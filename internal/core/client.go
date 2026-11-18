package core

import (
	"errors"
	"time"
)

const (
	ErrNotFound     = "item not found"
	ErrInvalidInput = "invalid input"
)

type Store interface {
	Create(item Todo) (Todo, error)
	Update(item Todo) (Todo, error)
	Get(id uint) (Todo, bool)
	List(includeDone bool) ([]Todo, error)
	Delete(id uint) error
}

type PecoClient struct {
	storage Store
}

func NewPecoClient(store Store) *PecoClient {
	return &PecoClient{
		storage: store,
	}
}

// Save creates or updates a Todo item in the storage.
// Items with non-zero IDs are treated as updates, while items with zero IDs are treated as new entries.
func (client *PecoClient) Save(item Todo) (Todo, error) {
	if item.Content == "" {
		return Todo{}, errors.New(ErrInvalidInput)
	}

	item.CreatedAt = item.CreatedAt.UTC()
	if item.Due != nil {
		due := item.Due.UTC()
		item.Due = &due
	}

	if item.Id == 0 {
		item.CreatedAt = time.Now().UTC()
		created, err := client.storage.Create(item)
		if err != nil {
			return Todo{}, err
		}

		return created, nil
	}

	existing, found := client.storage.Get(item.Id)

	if !found {
		return Todo{}, errors.New(ErrNotFound)
	}

	if !existing.CreatedAt.Equal(item.CreatedAt) {
		return Todo{}, errors.New(ErrInvalidInput)
	}

	return client.storage.Update(item)
}

// Get retrieves a Todo item by its ID from the storage.
// It returns an error if the item is not found.
func (client *PecoClient) Get(id uint) (Todo, error) {
	todo, ok := client.storage.Get(id)
	if !ok {
		return Todo{}, errors.New(ErrNotFound)
	}
	return todo, nil
}

// List returns pending items, or all items when includeDone is true.
// It returns an error if the underlying storage fails to retrieve the items.
func (client *PecoClient) List(includeDone bool) ([]Todo, error) {
	return client.storage.List(includeDone)
}

// Delete idempotently removes a Todo item from the storage by its ID.
func (client *PecoClient) Delete(id uint) error {
	return client.storage.Delete(id)
}
