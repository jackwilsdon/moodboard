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

type item struct {
	id    string
	image []byte
}

// Store represents an in-memory collection of moodboard items.
type Store struct {
	items []item
	mutex sync.RWMutex
}

// Create creates a new moodboard item in the collection.
func (s *Store) Create(img io.Reader) (string, error) {
	// Read the whole image into memory.
	buf, err := ioutil.ReadAll(img)

	if err != nil {
		return "", fmt.Errorf("failed to read image: %w", err)
	}

	// We're going to be modifying our items slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	id := uuid.New().String()

	s.items = append(s.items, item{
		id:    id,
		image: buf,
	})

	return id, nil
}

// All returns all moodboard items in the collection.
func (s *Store) All() ([]string, error) {
	// We're going to be reading from our items slice - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	if s.items == nil {
		return nil, nil
	}

	items := make([]string, len(s.items))

	// Extract the ID from each item.
	for i, item := range s.items {
		items[i] = item.id
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
		if s.items[i].id == id {
			return bytes.NewReader(s.items[i].image), nil
		}
	}

	return nil, moodboard.ErrNoSuchItem
}

// move moves a moodboard item before or after another one in the collection.
func (s *Store) move(id, targetID string, before bool) error {
	// We're going to be modifying our items slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	index := -1
	target := -1

	// Find the indexes of the items we're moving.
	for i, item := range s.items {
		if item.id == id {
			index = i
		}

		if item.id == targetID {
			target = i
		}

		// We can break early if we've found both indexes.
		if index != -1 && target != -1 {
			break
		}
	}

	// If either of the indexes is missing, return an error.
	if index == -1 || target == -1 {
		return moodboard.ErrNoSuchItem
	}

	item := s.items[index]

	if index < target {
		// If we're moving the item before the target and it's already before the target then we need to take
		// 1 off the target to ensure we don't insert it after the target.
		if before {
			target--
		}

		// index = 1
		// target = 4
		//
		// 0 1 2 3 4 5
		// -----------
		// A B C D E F
		// |  / / /  |
		// A C D E X F
		// | | | | | |
		// A C D E B F
		copy(s.items[index:], s.items[index+1:target+1])
	} else if index > target {
		// If we're moving the item after the target and it's already after the target then we need to add 1
		// to the target to ensure we don't insert it before the target.
		if !before {
			target++
		}

		// move([]string{"A", "B", "C", "D", "E", "F"}, 4, 1)
		//
		// 0 1 2 3 4 5
		// -----------
		// A B C D E F
		// |  \ \ \  |
		// A X B C D F
		// | | | | | |
		// A E B C D F
		copy(s.items[target+1:], s.items[target:index])
	}

	s.items[target] = item

	return nil
}

// MoveBefore moves a moodboard item before another one in the collection.
//
// This method will return moodboard.ErrNoSuchItem if items with either of the specified IDs do not exist.
func (s *Store) MoveBefore(id string, beforeID string) error {
	return s.move(id, beforeID, true)
}

// MoveAfter moves a moodboard item after another one in the collection.
//
// This method will return moodboard.ErrNoSuchItem if items with either of the specified IDs do not exist.
func (s *Store) MoveAfter(id string, afterID string) error {
	return s.move(id, afterID, false)
}

// Delete removes a moodboard item from the collection.
//
// This method will return moodboard.ErrNoSuchItem if an item with the specified ID does not exist.
func (s *Store) Delete(id string) error {
	// We're going to be modifying our items slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	remainingItems := make([]item, 0, len(s.items))

	// Only keep items which do not match the ID provided.
	for _, item := range s.items {
		if item.id != id {
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
