package file

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackwilsdon/moodboard"
	"io"
	"os"
	"path"
	"sync"
)

// Store represents an on-disk collection of moodboard items.
type Store struct {
	path  string
	mutex sync.RWMutex
}

// saveImage saves an image for a moodboard item in the collection.
func (s *Store) saveImage(img io.Reader, id string) (string, error) {
	f, err := os.OpenFile(path.Join(s.path, id), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)

	// If the file doesn't exist, try making the containing directory.
	if os.IsNotExist(err) {
		if err := os.MkdirAll(s.path, 0o777); err != nil {
			return "", fmt.Errorf("failed to create path: %w", err)
		}

		// Re-open the file now that we've created the containing directory.
		f, err = os.OpenFile(path.Join(s.path, id), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	}

	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	if _, err := io.Copy(f, img); err != nil {
		_ = f.Close()

		return "", fmt.Errorf("failed to write image: %w", err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close image: %w", err)
	}

	return f.Name(), nil
}

// Create creates a new moodboard item in the collection.
func (s *Store) Create(img io.Reader) (moodboard.Entry, error) {
	// We're going to be writing to disk - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	id := uuid.New().String()

	// Save the image - we can delete it later if something goes wrong.
	imgPath, err := s.saveImage(img, id)

	if err != nil {
		return moodboard.Entry{}, fmt.Errorf("failed to save image: %w", err)
	}

	// Open the file as R/W whilst optionally creating it if it doesn't exist.
	f, err := os.OpenFile(path.Join(s.path, "index.json"), os.O_RDWR|os.O_CREATE, 0o666)

	// If the file doesn't exist, try making the containing directory.
	if os.IsNotExist(err) {
		if err := os.MkdirAll(s.path, 0o777); err != nil {
			return moodboard.Entry{}, fmt.Errorf("failed to create path: %w", err)
		}

		// Re-open the file now that we've created the containing directory.
		f, err = os.OpenFile(path.Join(s.path, id), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o666)
	}

	if err != nil {
		_ = os.Remove(imgPath)

		return moodboard.Entry{}, fmt.Errorf("failed to open store: %w", err)
	}

	var entries []moodboard.Entry

	// Read the current entry list.
	if err = json.NewDecoder(f).Decode(&entries); err != nil && err != io.EOF {
		_ = os.Remove(imgPath)
		_ = f.Close()

		return moodboard.Entry{}, fmt.Errorf("failed to read store: %w", err)
	}

	// Jump back to the start of the file so that we can overwrite the existing entry list.
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		_ = os.Remove(imgPath)
		_ = f.Close()

		return moodboard.Entry{}, fmt.Errorf("failed to seek to start of file: %w", err)
	}

	entry := moodboard.Entry{ID: id}
	entries = append(entries, entry)

	// Write the new entry list.
	if err = json.NewEncoder(f).Encode(entries); err != nil {
		_ = os.Remove(imgPath)
		_ = f.Close()

		return moodboard.Entry{}, fmt.Errorf("failed to write store: %w", err)
	}

	// Close the file.
	if err = f.Close(); err != nil {
		_ = os.Remove(imgPath)

		return moodboard.Entry{}, fmt.Errorf("failed to close file: %w", err)
	}

	return entry, nil
}

// All returns all moodboard items in the collection.
func (s *Store) All() ([]moodboard.Entry, error) {
	// We're only going to be reading from the disk - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	// Open the file as read-only.
	f, err := os.Open(path.Join(s.path, "index.json"))

	// If the file doesn't exist then we can exit early (as there's nothing to list).
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to open store: %w", err)
	}

	var entries []moodboard.Entry

	// Read the current entry list.
	if err = json.NewDecoder(f).Decode(&entries); err != nil && err != io.EOF {
		_ = f.Close()

		return nil, fmt.Errorf("failed to read store: %w", err)
	}

	// We can ignore close errors here as we haven't written to the file.
	_ = f.Close()

	return entries, nil
}

// GetImage returns the image for the specified moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified ID does not exist.
func (s *Store) GetImage(id string) (io.Reader, error) {
	// We're only going to be reading from the disk - lock for reading.
	s.mutex.RLock()

	// Unlock once we're done.
	defer s.mutex.RUnlock()

	// Open the file as read-only.
	f, err := os.OpenFile(path.Join(s.path, "index.json"), os.O_RDONLY, 0)

	// If the file doesn't exist then we can exit early (as we don't have any images).
	if os.IsNotExist(err) {
		return nil, moodboard.ErrNoSuchEntry
	} else if err != nil {
		return nil, fmt.Errorf("failed to open store: %w", err)
	}

	var entries []moodboard.Entry

	// Read the current entry list.
	if err = json.NewDecoder(f).Decode(&entries); err != nil {
		_ = f.Close()

		// If it's an EOF error then we can ignore the error and exit early (as the file is empty).
		if err == io.EOF {
			return nil, moodboard.ErrNoSuchEntry
		}

		return nil, fmt.Errorf("failed to read store: %w", err)
	}

	// We can ignore close errors here as we haven't written to the file.
	_ = f.Close()

	var exists bool

	for _, entry := range entries {
		if entry.ID == id {
			exists = true
			break
		}
	}

	if !exists {
		return nil, moodboard.ErrNoSuchEntry
	}

	f, err = os.OpenFile(path.Join(s.path, id), os.O_RDONLY, 0)

	if os.IsNotExist(err) {
		return nil, moodboard.ErrNoSuchEntry
	} else if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	return f, nil
}

// Update updates a moodboard item in the collection.
//
// This method will return moodboard.ErrNoSuchEntry if an item with the specified ID does not exist.
func (s *Store) Update(entry moodboard.Entry) error {
	// We're going to be writing to disk - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	// Open the file as R/W.
	f, err := os.OpenFile(path.Join(s.path, "index.json"), os.O_RDWR, 0)

	// If the file doesn't exist then we can exit early (as there's nothing to update).
	if os.IsNotExist(err) {
		return moodboard.ErrNoSuchEntry
	} else if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	var entries []moodboard.Entry

	// Read the current entry list.
	if err = json.NewDecoder(f).Decode(&entries); err != nil {
		_ = f.Close()

		// If it's an EOF error then we can ignore the error and exit early (as the file is empty).
		if err == io.EOF {
			return moodboard.ErrNoSuchEntry
		}

		return fmt.Errorf("failed to read store: %w", err)
	}

	var exists bool

	// Replace the first entry we find with a matching ID.
	for i := range entries {
		if entries[i].ID == entry.ID {
			entries[i] = entry
			exists = true

			break
		}
	}

	// Make sure we actually updated an entry.
	if !exists {
		_ = f.Close()

		return moodboard.ErrNoSuchEntry
	}

	// Jump back to the start of the file so that we can overwrite the existing entry list.
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to seek to start of file: %w", err)
	}

	// Write the new entry list.
	if err = json.NewEncoder(f).Encode(entries); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to write store: %w", err)
	}

	// Close the file.
	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// Delete removes a moodboard item from the collection.
//
// This method will return moodboard.ErrNoSuchEntry an item with the specified ID does not exist.
func (s *Store) Delete(id string) error {
	// We're going to be writing to disk - lock for writing.
	s.mutex.Lock()

	// Unlock once we're done.
	defer s.mutex.Unlock()

	// Open the file as R/W.
	f, err := os.OpenFile(path.Join(s.path, "index.json"), os.O_RDWR, 0)

	// If the file doesn't exist then we can exit early (as there's nothing to delete).
	if os.IsNotExist(err) {
		return moodboard.ErrNoSuchEntry
	} else if err != nil {
		return fmt.Errorf("failed to open store: %w", err)
	}

	var entries []moodboard.Entry

	// Read the current entry list.
	if err = json.NewDecoder(f).Decode(&entries); err != nil {
		_ = f.Close()

		// If it's an EOF error then we can ignore the error and exit early (as the file is empty).
		if err == io.EOF {
			return moodboard.ErrNoSuchEntry
		}

		return fmt.Errorf("failed to read store: %w", err)
	}

	remainingEntries := make([]moodboard.Entry, 0, len(entries))

	// Only keep entries which do not match the ID provided.
	for _, entry := range entries {
		if entry.ID != id {
			remainingEntries = append(remainingEntries, entry)
		}
	}

	// If the number of entries is the same then we haven't found anything to delete.
	if len(entries) == len(remainingEntries) {
		_ = f.Close()

		return moodboard.ErrNoSuchEntry
	}

	// Jump back to the start of the file so that we can overwrite the existing entry list.
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to seek to start of file: %w", err)
	}

	// Write the new entry list.
	if err = json.NewEncoder(f).Encode(remainingEntries); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to write store: %w", err)
	}

	// Work out our current position so that we can truncate the remainder of the file.
	pos, err := f.Seek(0, io.SeekCurrent)

	if err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to find position in file: %w", err)
	}

	// Truncate the remainder of the file.
	if err = f.Truncate(pos); err != nil {
		_ = f.Close()

		return fmt.Errorf("failed to truncate file: %w", err)
	}

	// Close the file.
	if err = f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// NewStore creates a new moodboard collection, backed by the directory at the specified path.
func NewStore(path string) *Store {
	return &Store{path: path}
}
