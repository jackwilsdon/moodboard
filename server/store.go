package moodboard

import (
	"errors"
	"io"
)

// ErrNoSuchItem indicates that an item does not exist.
var ErrNoSuchItem = errors.New("no such item")

// Item represents a single moodboard item.
type Item struct {
	ID string `json:"id"`

	// A number in the range [0, 1].
	X float32 `json:"x"`

	// A number in the range [0, 1].
	Y float32 `json:"y"`

	// A number in the range [0, 1].
	Width float32 `json:"width"`
}

// Store represents a collection of moodboard items.
type Store interface {
	// Create creates a new moodboard item in the collection.
	Create(io.Reader) (Item, error)

	// All returns all moodboard items in the collection.
	All() ([]Item, error)

	// GetImage returns the image for the specified moodboard item in the collection.
	//
	// Note that the reader returned by this method may be an io.ReadCloser.
	//
	// This method will return ErrNoSuchItem if an item with the specified ID does not exist.
	GetImage(id string) (io.Reader, error)

	// Update updates a moodboard item in the collection.
	//
	// This method will return ErrNoSuchItem if an item with the specified ID does not exist.
	Update(item Item) error

	// Delete removes a moodboard item from the collection.
	//
	// This method will return ErrNoSuchItem if an item with the specified ID does not exist.
	Delete(id string) error
}
