package file_test

import (
	"github.com/jackwilsdon/moodboard"
	"github.com/jackwilsdon/moodboard/file"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

// newStore creates a new moodboard store for testing.
//
// A temporary directory is used to back the store, which is cleaned up once the test and all its subtests complete.
func newStore(t *testing.T) *file.Store {
	dir, err := ioutil.TempDir("", "store")

	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}

	// Delete the directory at the end of the test.
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return file.NewStore(path.Join(dir, "store.json"))
}

func TestStoreCreate(t *testing.T) {
	s := newStore(t)

	entry, err := s.Create()

	if err != nil {
		t.Fatalf("expected error to be nil but got %q", err)
	}

	if entry.X != 0 {
		t.Errorf("expected entry.X to be 0 but got %v", entry.X)
	}

	if entry.Y != 0 {
		t.Errorf("expected entry.Y to be 0 but got %v", entry.Y)
	}

	if entry.Width != 0 {
		t.Errorf("expected entry.Width to be 0 but got %v", entry.Width)
	}
}

func TestStoreUpdate(t *testing.T) {
	cs := []struct {
		name   string
		create int
		update int
		entry  moodboard.Entry
		err    error
	}{
		{
			name:   "update",
			create: 3,
			update: 0,
			entry: moodboard.Entry{
				X:     0.1,
				Y:     0.2,
				Width: 0.3,
			},
		},
		{
			name:   "update nonexistent",
			create: 1,
			update: -1,
			err:    moodboard.ErrNoSuchEntry,
		},
		{
			name:   "update empty",
			update: -1,
			err:    moodboard.ErrNoSuchEntry,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)
			entries := make([]moodboard.Entry, c.create)

			for i := 0; i < c.create; i++ {
				entry, err := s.Create()

				if err != nil {
					t.Fatalf("failed to create entry: %v", err)
				}

				entries[i] = entry
			}

			entry := c.entry

			if c.update != -1 {
				entry.ID = entries[c.update].ID
				entries[c.update] = entry
			}

			switch err := s.Update(entry); {
			case err != nil && c.err == nil:
				t.Fatalf("expected error to be nil but got %q", err)
			case err == nil && c.err != nil:
				t.Fatalf("expected error to be %q but got nil", c.err)
			case err != c.err:
				t.Fatalf("expected error to be %q but got %q", c.err, err)
			}

			all, err := s.All()

			if err != nil {
				t.Fatalf("failed to get store contents: %v", err)
			}

			if len(all) != len(entries) {
				verb := "entries"

				if len(entries) == 1 {
					verb = "entry"
				}

				t.Fatalf("expected to get %d %s but got %d", len(entries), verb, len(all))
			}

			for i := range all {
				if all[i].ID != entries[i].ID {
					t.Errorf("expected all[%d].ID to be %v but got %v", i, entries[i].ID, all[i].ID)
				}

				if all[i].X != entries[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, entries[i].X, all[i].X)
				}

				if all[i].Y != entries[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, entries[i].Y, all[i].Y)
				}

				if all[i].Width != entries[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, entries[i].Width, all[i].Width)
				}
			}
		})
	}
}

func TestStoreDelete(t *testing.T) {
	cs := []struct {
		name   string
		create int
		delete int
		err    error
	}{
		{
			name:   "delete",
			create: 3,
			delete: 0,
		},
		{
			name:   "delete nonexistent",
			create: 1,
			delete: -1,
			err:    moodboard.ErrNoSuchEntry,
		}, {
			name:   "delete empty",
			delete: -1,
			err:    moodboard.ErrNoSuchEntry,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)
			entries := make([]moodboard.Entry, c.create)

			for i := 0; i < c.create; i++ {
				entry, err := s.Create()

				if err != nil {
					t.Fatalf("failed to create entry: %v", err)
				}

				entries[i] = entry
			}

			var id string

			if c.delete != -1 {
				id = entries[c.delete].ID

				// Move all entries after the deleted one left.
				copy(entries[c.delete:], entries[c.delete+1:])

				// Remove the last (now duplicated) element.
				entries = entries[:len(entries)-1]
			}

			switch err := s.Delete(id); {
			case err != nil && c.err == nil:
				t.Fatalf("expected error to be nil but got %q", err)
			case err == nil && c.err != nil:
				t.Fatalf("expected error to be %q but got nil", c.err)
			case err != c.err:
				t.Fatalf("expected error to be %q but got %q", c.err, err)
			}

			all, err := s.All()

			if err != nil {
				t.Fatalf("failed to get store contents: %v", err)
			}

			if len(all) != len(entries) {
				verb := "entries"

				if len(entries) == 1 {
					verb = "entry"
				}

				t.Fatalf("expected to get %d %s but got %d", len(entries), verb, len(all))
			}

			for i := range all {
				if all[i].ID != entries[i].ID {
					t.Errorf("expected all[%d].ID to be %v but got %v", i, entries[i].ID, all[i].ID)
				}

				if all[i].X != entries[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, entries[i].X, all[i].X)
				}

				if all[i].Y != entries[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, entries[i].Y, all[i].Y)
				}

				if all[i].Width != entries[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, entries[i].Width, all[i].Width)
				}
			}
		})
	}
}
