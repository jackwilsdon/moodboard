package file_test

import (
	"bytes"
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
// A temporary directory is used to back the store, which is cleaned up once the test and all its subtests complete.
func newStore(t *testing.T) *file.Store {
	dir, err := ioutil.TempDir("", "")

	if err != nil {
		t.Fatalf("failed to create temporary directory: %v", err)
	}

	// Delete the directory at the end of the test.
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return file.NewStore(path.Join(dir, "data"))
}

func TestStoreCreate(t *testing.T) {
	s := newStore(t)

	item, err := s.Create(bytes.NewReader(nil))

	if err != nil {
		t.Fatalf("expected error to be nil but got %q", err)
	}

	if item.X != 0 {
		t.Errorf("expected item.X to be 0 but got %v", item.X)
	}

	if item.Y != 0 {
		t.Errorf("expected item.Y to be 0 but got %v", item.Y)
	}

	if item.Width != 0 {
		t.Errorf("expected item.Width to be 0 but got %v", item.Width)
	}
}

func TestStoreGetImage(t *testing.T) {
	cs := []struct {
		name   string
		create int
		get    int
		err    error
	}{
		{
			name:   "get",
			create: 3,
			get:    0,
			err:    nil,
		},
		{
			name:   "get nonexistent",
			create: 0,
			get:    -1,
			err:    moodboard.ErrNoSuchItem,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)

			var id string
			var expectedImg []byte

			for i := 0; i < c.create; i++ {
				img := []byte(fmt.Sprintf("image %d", i))
				item, err := s.Create(bytes.NewReader(img))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				if i == c.get {
					id = item.ID
					expectedImg = img
				}
			}

			img, err := s.GetImage(id)

			switch {
			case err != nil && c.err == nil:
				t.Fatalf("expected error to be nil but got %q", err)
			case err == nil && c.err != nil:
				t.Fatalf("expected error to be %q but got nil", c.err)
			case err != c.err:
				t.Fatalf("expected error to be %q but got %q", c.err, err)
			}

			if err == nil {
				imgBytes, err := ioutil.ReadAll(img)

				if err != nil {
					t.Fatalf("failed to read returned image: %v", err)
				}

				if !bytes.Equal(imgBytes, expectedImg) {
					t.Fatalf("expected returned reader to read %q but got %q", expectedImg, imgBytes)
				}
			}
		})
	}
}

func TestStoreUpdate(t *testing.T) {
	cs := []struct {
		name   string
		create int
		update int
		item   moodboard.Item
		err    error
	}{
		{
			name:   "update",
			create: 3,
			update: 0,
			item: moodboard.Item{
				X:     0.1,
				Y:     0.2,
				Width: 0.3,
			},
		},
		{
			name:   "update nonexistent",
			create: 1,
			update: -1,
			err:    moodboard.ErrNoSuchItem,
		},
		{
			name:   "update empty",
			update: -1,
			err:    moodboard.ErrNoSuchItem,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)
			items := make([]moodboard.Item, c.create)

			for i := 0; i < c.create; i++ {
				item, err := s.Create(bytes.NewReader(nil))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				items[i] = item
			}

			item := c.item

			if c.update != -1 {
				item.ID = items[c.update].ID
				items[c.update] = item
			}

			switch err := s.Update(item); {
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

			if len(all) != len(items) {
				verb := "items"

				if len(items) == 1 {
					verb = "item"
				}

				t.Fatalf("expected to get %d %s but got %d", len(items), verb, len(all))
			}

			for i := range all {
				if all[i].ID != items[i].ID {
					t.Errorf("expected all[%d].ID to be %v but got %v", i, items[i].ID, all[i].ID)
				}

				if all[i].X != items[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, items[i].X, all[i].X)
				}

				if all[i].Y != items[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, items[i].Y, all[i].Y)
				}

				if all[i].Width != items[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, items[i].Width, all[i].Width)
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
			err:    moodboard.ErrNoSuchItem,
		}, {
			name:   "delete empty",
			delete: -1,
			err:    moodboard.ErrNoSuchItem,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := newStore(t)
			items := make([]moodboard.Item, c.create)

			for i := 0; i < c.create; i++ {
				item, err := s.Create(bytes.NewReader(nil))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				items[i] = item
			}

			var id string

			if c.delete != -1 {
				id = items[c.delete].ID

				// Move all items after the deleted one left.
				copy(items[c.delete:], items[c.delete+1:])

				// Remove the last (now duplicated) element.
				items = items[:len(items)-1]
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

			if len(all) != len(items) {
				verb := "items"

				if len(items) == 1 {
					verb = "item"
				}

				t.Fatalf("expected to get %d %s but got %d", len(items), verb, len(all))
			}

			for i := range all {
				if all[i].ID != items[i].ID {
					t.Errorf("expected all[%d].ID to be %v but got %v", i, items[i].ID, all[i].ID)
				}

				if all[i].X != items[i].X {
					t.Errorf("expected all[%d].X to be %v but got %v", i, items[i].X, all[i].X)
				}

				if all[i].Y != items[i].Y {
					t.Errorf("expected all[%d].Y to be %v but got %v", i, items[i].Y, all[i].Y)
				}

				if all[i].Width != items[i].Width {
					t.Errorf("expected all[%d].Width to be %v but got %v", i, items[i].Width, all[i].Width)
				}
			}
		})
	}
}
