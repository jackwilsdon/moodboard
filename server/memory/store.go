package memory

import (
	"github.com/jackwilsdon/moodboard"
	"sync"
)

// Store represents an in-memory collection of moodboard items.
type Store struct {
	entries []moodboard.Entry
	mutex   sync.RWMutex
}

// Insert adds a new moodboard item to the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified URL already exists.
func (s *Store) Insert(entry moodboard.Entry) error {
	// We're going to be modifying our entries slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	// Ensure we're not inserting a duplicate.
	for _, existing := range s.entries {
		if existing.URL == entry.URL {
			return moodboard.ErrDuplicateURL
		}
	}

	s.entries = append(s.entries, entry)

	return nil
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

	// Make a copy of our entries slice.
	entries := make([]moodboard.Entry, len(s.entries))
	copy(entries, s.entries)

	return entries, nil
}

// Update updates a moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified URL does not exist.
func (s *Store) Update(entry moodboard.Entry) error {
	// We're going to be modifying our entries slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	// Replace the first entry we find with a matching URL.
	for i := range s.entries {
		if s.entries[i].URL == entry.URL {
			s.entries[i] = entry
			return nil
		}
	}

	return moodboard.ErrNoSuchEntry
}

// Delete removes a moodboard item from the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified URL does not exist.
func (s *Store) Delete(url string) error {
	// We're going to be modifying our entries slice - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	remainingEntries := make([]moodboard.Entry, 0, len(s.entries))

	// Only keep entries which do not match the URL provided.
	for _, entry := range s.entries {
		if entry.URL != url {
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
func NewStore(entries []moodboard.Entry) *Store {
	s := &Store{}

	// If we were given some entries then copy them into the store.
	if entries != nil {
		s.entries = make([]moodboard.Entry, len(entries))
		copy(s.entries, entries)
	}

	return s
}
