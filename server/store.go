package moodboard

import (
	"errors"
	"io"
)

// ErrNoSuchItem indicates that an item does not exist.
var ErrNoSuchItem = errors.New("no such item")

// Store represents a collection of moodboard items.
type Store interface {
	// Create creates a new moodboard item in the collection.
	Create(io.Reader) (string, error)

	// All returns all moodboard items in the collection.
	All() ([]string, error)

	// GetImage returns the image for the specified moodboard item in the collection.
	//
	// Note that the reader returned by this method may be an io.ReadCloser.
	//
	// This method will return ErrNoSuchItem if an item with the specified ID does not exist.
	GetImage(id string) (io.Reader, error)

	// MoveBefore moves a moodboard item before another one in the collection.
	//
	// This method will return ErrNoSuchItem if items with either of the specified IDs do not exist.
	MoveBefore(id, beforeID string) error

	// MoveAfter moves a moodboard item after another one in the collection.
	//
	// This method will return ErrNoSuchItem if items with either of the specified IDs do not exist.
	MoveAfter(id, afterID string) error

	// Delete removes a moodboard item from the collection.
	//
	// This method will return ErrNoSuchItem if an item with the specified ID does not exist.
	Delete(id string) error
}
