package file_test

import (
	"fmt"
	"github.com/jackwilsdon/moodboard"
	"github.com/jackwilsdon/moodboard/file"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

// newStore creates a new moodboard store for testing.
//
// A temporary file is used to back the store, which is cleaned up once the test and all its subtests complete.
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

func TestStoreInsert(t *testing.T) {
	type insertOp struct {
		entry moodboard.Entry
		err   error
	}

	cs := []struct {
		name    string
		inserts []insertOp
		all     []moodboard.Entry
	}{
		{
			name: "insert",
			inserts: []insertOp{
				{
					entry: moodboard.Entry{
						URL:   "https://example.com/1",
						X:     0.1,
						Y:     0.2,
						Width: 0.3,
					},
				},
				{
					entry: moodboard.Entry{
						URL:   "https://example.com/2",
						X:     0.4,
						Y:     0.5,
						Width: 0.6,
					},
				},
			},
			all: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
				{
					URL:   "https://example.com/2",
					X:     0.4,
					Y:     0.5,
					Width: 0.6,
				},
			},
		},
		{
			name: "insert duplicate",
			inserts: []insertOp{
				{
					entry: moodboard.Entry{
						URL:   "https://example.com",
						X:     0.1,
						Y:     0.2,
						Width: 0.3,
					},
				},
				{
					entry: moodboard.Entry{
						URL:   "https://example.com",
						X:     0.4,
						Y:     0.5,
						Width: 0.6,
					},
					err: moodboard.ErrDuplicateURL,
				},
			},
			all: []moodboard.Entry{
				{
					URL:   "https://example.com",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
			},
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)

			for i, op := range c.inserts {
				t.Run(fmt.Sprintf("insert %d", i), func(t *testing.T) {
					switch err := s.Insert(op.entry); {
					case err != nil && op.err == nil:
						t.Fatalf("expected error to be nil but got %q", err)
					case err == nil && op.err != nil:
						t.Fatalf("expected error to be %q but got nil", op.err)
					case err != op.err:
						t.Fatalf("expected error to be %q but got %q", op.err, err)
					}
				})
			}

			all, err := s.All()

			if err != nil {
				t.Fatalf("failed to get store contents: %v", err)
			}

			if len(all) != len(c.all) {
				verb := "entries"

				if len(c.all) == 1 {
					verb = "entry"
				}

				t.Fatalf("expected to get %d %s but got %d", len(c.all), verb, len(all))
			}

			for i := range all {
				if all[i].URL != c.all[i].URL {
					t.Errorf("expected all[%d].URL to be %q but got %q", i, c.all[i].URL, all[i].URL)
				}

				if all[i].X != c.all[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, c.all[i].X, all[i].X)
				}

				if all[i].Y != c.all[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, c.all[i].Y, all[i].Y)
				}

				if all[i].Width != c.all[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, c.all[i].Width, all[i].Width)
				}
			}
		})
	}
}

func TestStoreUpdate(t *testing.T) {
	type updateOp struct {
		entry moodboard.Entry
		err   error
	}

	cs := []struct {
		name    string
		inserts []moodboard.Entry
		updates []updateOp
		all     []moodboard.Entry
	}{
		{
			name: "update",
			inserts: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
				{
					URL:   "https://example.com/2",
					X:     0.4,
					Y:     0.5,
					Width: 0.6,
				},
				{
					URL:   "https://example.com/3",
					X:     0.7,
					Y:     0.8,
					Width: 0.9,
				},
			},
			updates: []updateOp{
				{
					entry: moodboard.Entry{
						URL:   "https://example.com/1",
						X:     0.7,
						Y:     0.8,
						Width: 0.9,
					},
				},
				{
					entry: moodboard.Entry{
						URL:   "https://example.com/3",
						X:     0.1,
						Y:     0.2,
						Width: 0.3,
					},
				},
			},
			all: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.7,
					Y:     0.8,
					Width: 0.9,
				},
				{
					URL:   "https://example.com/2",
					X:     0.4,
					Y:     0.5,
					Width: 0.6,
				},
				{
					URL:   "https://example.com/3",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
			},
		},
		{
			name: "update nonexistent",
			inserts: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
			},
			updates: []updateOp{
				{
					entry: moodboard.Entry{
						URL:   "https://example.com/2",
						X:     0.4,
						Y:     0.5,
						Width: 0.6,
					},
					err: moodboard.ErrNoSuchEntry,
				},
			},
			all: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
			},
		},
		{
			name: "update empty",
			updates: []updateOp{
				{
					entry: moodboard.Entry{
						URL:   "https://example.com/2",
						X:     0.1,
						Y:     0.2,
						Width: 0.3,
					},
					err: moodboard.ErrNoSuchEntry,
				},
			},
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)

			for i, e := range c.inserts {
				err := s.Insert(e)

				if err != nil {
					t.Fatalf("failed to insert entry %d: %v", i, err)
				}
			}

			for i, op := range c.updates {
				t.Run(fmt.Sprintf("update %d", i), func(t *testing.T) {
					switch err := s.Update(op.entry); {
					case err != nil && op.err == nil:
						t.Fatalf("expected error to be nil but got %q", err)
					case err == nil && op.err != nil:
						t.Fatalf("expected error to be %q but got nil", op.err)
					case err != op.err:
						t.Fatalf("expected error to be %q but got %q", op.err, err)
					}
				})
			}

			all, err := s.All()

			if err != nil {
				t.Fatalf("failed to get store contents: %v", err)
			}

			if len(all) != len(c.all) {
				verb := "entries"

				if len(c.all) == 1 {
					verb = "entry"
				}

				t.Fatalf("expected to get %d %s but got %d", len(c.all), verb, len(all))
			}

			for i := range all {
				if all[i].URL != c.all[i].URL {
					t.Errorf("expected all[%d].URL to be %q but got %q", i, c.all[i].URL, all[i].URL)
				}

				if all[i].X != c.all[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, c.all[i].X, all[i].X)
				}

				if all[i].Y != c.all[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, c.all[i].Y, all[i].Y)
				}

				if all[i].Width != c.all[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, c.all[i].Width, all[i].Width)
				}
			}
		})
	}
}

func TestStoreUpdateEmptyFile(t *testing.T) {
	f, err := ioutil.TempFile("", "store.*.json")

	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	// We can close the file immediately as the store will open it itself.
	_ = f.Close()

	// Delete the file at the end of the test.
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})

	s := file.NewStore(f.Name())

	if err := s.Update(moodboard.Entry{}); err != moodboard.ErrNoSuchEntry {
		if err == nil {
			t.Fatalf("expected error to be %q but got nil", moodboard.ErrNoSuchEntry)
		} else {
			t.Fatalf("expected error to be %q but got %q", moodboard.ErrNoSuchEntry, err)
		}
	}
}

func TestStoreDelete(t *testing.T) {
	type deleteOp struct {
		url string
		err error
	}

	cs := []struct {
		name    string
		inserts []moodboard.Entry
		deletes []deleteOp
		all     []moodboard.Entry
	}{
		{
			name: "delete",
			inserts: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
				{
					URL:   "https://example.com/2",
					X:     0.4,
					Y:     0.5,
					Width: 0.6,
				},
				{
					URL:   "https://example.com/3",
					X:     0.7,
					Y:     0.8,
					Width: 0.9,
				},
			},
			deletes: []deleteOp{
				{
					url: "https://example.com/1",
				},
				{
					url: "https://example.com/3",
				},
			},
			all: []moodboard.Entry{
				{
					URL:   "https://example.com/2",
					X:     0.4,
					Y:     0.5,
					Width: 0.6,
				},
			},
		},
		{
			name: "delete nonexistent",
			inserts: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
			},
			deletes: []deleteOp{
				{
					url: "https://example.com/2",
					err: moodboard.ErrNoSuchEntry,
				},
			},
			all: []moodboard.Entry{
				{
					URL:   "https://example.com/1",
					X:     0.1,
					Y:     0.2,
					Width: 0.3,
				},
			},
		}, {
			name: "delete empty",
			deletes: []deleteOp{
				{
					url: "https://example.com/2",
					err: moodboard.ErrNoSuchEntry,
				},
			},
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)

			for i, e := range c.inserts {
				err := s.Insert(e)

				if err != nil {
					t.Fatalf("failed to insert entry %d: %v", i, err)
				}
			}

			for i, op := range c.deletes {
				t.Run(fmt.Sprintf("delete %d", i), func(t *testing.T) {
					switch err := s.Delete(op.url); {
					case err != nil && op.err == nil:
						t.Fatalf("expected error to be nil but got %q", err)
					case err == nil && op.err != nil:
						t.Fatalf("expected error to be %q but got nil", op.err)
					case err != op.err:
						t.Fatalf("expected error to be %q but got %q", op.err, err)
					}
				})
			}

			all, err := s.All()

			if err != nil {
				t.Fatalf("failed to get store contents: %v", err)
			}

			if len(all) != len(c.all) {
				verb := "entries"

				if len(c.all) == 1 {
					verb = "entry"
				}

				t.Fatalf("expected to get %d %s but got %d", len(c.all), verb, len(all))
			}

			for i := range all {
				if all[i].URL != c.all[i].URL {
					t.Errorf("expected all[%d].URL to be %q but got %q", i, c.all[i].URL, all[i].URL)
				}

				if all[i].X != c.all[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, c.all[i].X, all[i].X)
				}

				if all[i].Y != c.all[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, c.all[i].Y, all[i].Y)
				}

				if all[i].Width != c.all[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, c.all[i].Width, all[i].Width)
				}
			}
		})
	}
}

func TestStoreDeleteEmptyFile(t *testing.T) {
	f, err := ioutil.TempFile("", "store.*.json")

	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}

	// We can close the file immediately as the store will open it itself.
	_ = f.Close()

	// Delete the file at the end of the test.
	t.Cleanup(func() {
		_ = os.Remove(f.Name())
	})

	s := file.NewStore(f.Name())

	if err := s.Delete("https://example.com"); err != moodboard.ErrNoSuchEntry {
		if err == nil {
			t.Fatalf("expected error to be %q but got nil", moodboard.ErrNoSuchEntry)
		} else {
			t.Fatalf("expected error to be %q but got %q", moodboard.ErrNoSuchEntry, err)
		}
	}
}
