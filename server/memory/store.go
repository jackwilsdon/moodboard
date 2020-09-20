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

type imageEntry struct {
	entry moodboard.Entry
	image []byte
}

// Store represents an in-memory collection of moodboard items.
type Store struct {
	entries []imageEntry
	mutex   sync.RWMutex
}

// Create creates a new moodboard item in the collection.
func (s *Store) Create(img io.Reader) (moodboard.Entry, error) {
	// Read the whole image into memory.
	buf, err := ioutil.ReadAll(img)

	if err != nil {
		return moodboard.Entry{}, fmt.Errorf("failed to read image: %w", err)
	}

	// We're going to be modifying our entries slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	entry := moodboard.Entry{
		ID: uuid.New().String(),
	}

	s.entries = append(s.entries, imageEntry{
		entry: entry,
		image: buf,
	})

	return entry, nil
}

// All returns all moodboard items in the collection.
func (s *Store) All() ([]moodboard.Entry, error) {
	// We're going to be reading from our entries slice - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	if s.entries == nil {
		return nil, nil
	}

	entries := make([]moodboard.Entry, len(s.entries))

	// Extract the moodboard entry from each item.
	for i, entry := range s.entries {
		entries[i] = entry.entry
	}

	return entries, nil
}

// GetImage returns the image for the specified moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified ID does not exist.
func (s *Store) GetImage(id string) (io.Reader, error) {
	// We're going to be reading from our entries slice - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	// Return the image for the first entry we find with a matching ID.
	for i := range s.entries {
		if s.entries[i].entry.ID == id {
			return bytes.NewReader(s.entries[i].image), nil
		}
	}

	return nil, moodboard.ErrNoSuchEntry
}

// Update updates a moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified ID does not exist.
func (s *Store) Update(entry moodboard.Entry) error {
	// We're going to be modifying our entries slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	// Replace the first entry we find with a matching ID.
	for i := range s.entries {
		if s.entries[i].entry.ID == entry.ID {
			s.entries[i].entry = entry
			return nil
		}
	}

	return moodboard.ErrNoSuchEntry
}

// Delete removes a moodboard item from the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified ID does not exist.
func (s *Store) Delete(id string) error {
	// We're going to be modifying our entries slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	remainingEntries := make([]imageEntry, 0, len(s.entries))

	// Only keep entries which do not match the ID provided.
	for _, entry := range s.entries {
		if entry.entry.ID != id {
			remainingEntries = append(remainingEntries, entry)
		}
	}

	// If the number of entries is the same then we haven't found anything to delete.
	if len(s.entries) == len(remainingEntries) {
		return moodboard.ErrNoSuchEntry
	}

	s.entries = remainingEntries

	return nil
}

// NewStore creates a new in-memory moodboard collection.
func NewStore() *Store {
	return &Store{}
}
