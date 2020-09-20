package moodboard

import (
	"errors"
	"io"
)

// ErrNoSuchEntry indicates that an entry does not exist.
var ErrNoSuchEntry = errors.New("no such entry")

// Entry represents a single moodboard item.
type Entry struct {
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
	Create(io.Reader) (Entry, error)

	// All returns all moodboard items in the collection.
	All() ([]Entry, error)

	// GetImage returns the image for the specified moodboard item in the collection.
	//
	// Note that the reader returned by this method may be an io.ReadCloser.
	//
	// This method will return ErrNoSuchEntry if an item with the specified ID does not exist.
	GetImage(id string) (io.Reader, error)

	// Update updates a moodboard item in the collection.
	//
	// This method will return ErrNoSuchEntry if an item with the specified ID does not exist.
	Update(entry Entry) error

	// Delete removes a moodboard item from the collection.
	//
	// This method will return ErrNoSuchEntry if an item with the specified ID does not exist.
	Delete(id string) error
}
