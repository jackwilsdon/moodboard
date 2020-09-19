package moodboard

import "errors"

// ErrDuplicateURL indicates that an entry has a duplicate URL.
var ErrDuplicateURL = errors.New("duplicate URL")

// ErrNoSuchEntry indicates that an entry does not exist.
var ErrNoSuchEntry = errors.New("no such entry")

// Entry represents a single moodboard item.
type Entry struct {
	URL string `json:"url"`

	// A number in the range [0, 1].
	X float32 `json:"x"`

	// A number in the range [0, 1].
	Y float32 `json:"y"`

	// A number in the range [0, 1].
	Width float32 `json:"width"`
}

// Store represents a collection of moodboard items.
type Store interface {
	// Insert adds a new moodboard item to the collection.
	//
	// This method will return ErrNoSuchEntry if an item with the specified URL already exists.
	Insert(Entry) error

	// All returns all moodboard items in the collection.
	All() ([]Entry, error)

	// Update updates a moodboard item in the collection.
	//
	// This method will return ErrNoSuchEntry if an item with the specified URL does not exist.
	Update(entry Entry) error

	// Delete removes a moodboard item from the collection.
	//
	// This method will return ErrNoSuchEntry if an item with the specified URL does not exist.
	Delete(url string) error
}
