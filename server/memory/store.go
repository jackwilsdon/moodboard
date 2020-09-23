package memory

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackwilsdon/moodboard"
	"io"
	"io/ioutil"
	"sync"
)

type imageItem struct {
	item  moodboard.Item
	image []byte
}

// Store represents an in-memory collection of moodboard items.
type Store struct {
	items []imageItem
	mutex sync.RWMutex
}

// Create creates a new moodboard item in the collection.
func (s *Store) Create(img io.Reader) (moodboard.Item, error) {
	// Read the whole image into memory.
	buf, err := ioutil.ReadAll(img)

	if err != nil {
		return moodboard.Item{}, fmt.Errorf("failed to read image: %w", err)
	}

	// We're going to be modifying our items slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	item := moodboard.Item{
		ID: uuid.New().String(),
	}

	s.items = append(s.items, imageItem{
		item:  item,
		image: buf,
	})

	return item, nil
}

// All returns all moodboard items in the collection.
func (s *Store) All() ([]moodboard.Item, error) {
	// We're going to be reading from our items slice - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	if s.items == nil {
		return nil, nil
	}

	items := make([]moodboard.Item, len(s.items))

	// Extract the moodboard item from each item.
	for i, item := range s.items {
		items[i] = item.item
	}

	return items, nil
}

// GetImage returns the image for the specified moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchItem if an item with the specified ID does not exist.
func (s *Store) GetImage(id string) (io.Reader, error) {
	// We're going to be reading from our items slice - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	// Return the image for the first item we find with a matching ID.
	for i := range s.items {
		if s.items[i].item.ID == id {
			return bytes.NewReader(s.items[i].image), nil
		}
	}

	return nil, moodboard.ErrNoSuchItem
}

// Update updates a moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchItem if an item with the specified ID does not exist.
func (s *Store) Update(item moodboard.Item) error {
	// We're going to be modifying our items slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	// Replace the first item we find with a matching ID.
	for i := range s.items {
		if s.items[i].item.ID == item.ID {
			s.items[i].item = item
			return nil
		}
	}

	return moodboard.ErrNoSuchItem
}

// Delete removes a moodboard item from the collection.
//
// This method will return moodboard.ErrNoSuchItem if an item with the specified ID does not exist.
func (s *Store) Delete(id string) error {
	// We're going to be modifying our items slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	remainingItems := make([]imageItem, 0, len(s.items))

	// Only keep items which do not match the ID provided.
	for _, item := range s.items {
		if item.item.ID != id {
			remainingItems = append(remainingItems, item)
		}
	}

	// If the number of items is the same then we haven't found anything to delete.
	if len(s.items) == len(remainingItems) {
		return moodboard.ErrNoSuchItem
	}

	s.items = remainingItems

	return nil
}

// NewStore creates a new in-memory moodboard collection.
func NewStore() *Store {
	return &Store{}
}
