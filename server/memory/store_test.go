package memory_test

import (
	"bytes"
	"fmt"
	"github.com/jackwilsdon/moodboard"
	"github.com/jackwilsdon/moodboard/memory"
	"io/ioutil"
	"testing"
)

func TestStoreCreate(t *testing.T) {
	s := memory.NewStore()

	firstID, err := s.Create(bytes.NewReader(nil))

	if err != nil {
		t.Fatalf("expected error to be nil but got %q", err)
	}

	secondID, err := s.Create(bytes.NewReader(nil))

	if err != nil {
		t.Fatalf("expected error to be nil but got %q", err)
	}

	all, err := s.All()

	if err != nil {
		t.Fatalf("failed to get store contents: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("expected to get 2 items but got %d", len(all))
	}

	if all[0] != firstID || all[1] != secondID {
		t.Fatalf("expected all to be [%q, %q] but got [%q, %q]", firstID, secondID, all[0], all[1])
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
			s := memory.NewStore()

			var targetID string
			var expectedImg []byte

			for i := 0; i < c.create; i++ {
				img := []byte(fmt.Sprintf("image %d", i))
				id, err := s.Create(bytes.NewReader(img))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				if i == c.get {
					targetID = id
					expectedImg = img
				}
			}

			img, err := s.GetImage(targetID)

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

func TestStoreMoveBefore(t *testing.T) {
	cs := []struct {
		name   string
		index  int
		before int
		order  []int
		err    error
	}{
		{
			name:   "move first before first",
			index:  0,
			before: 0,
			order:  []int{0, 1, 2},
		},
		{
			name:   "move first before second",
			index:  0,
			before: 1,
			order:  []int{0, 1, 2},
		},
		{
			name:   "move first before third",
			index:  0,
			before: 2,
			order:  []int{1, 0, 2},
		},
		{
			name:   "move second before first",
			index:  1,
			before: 0,
			order:  []int{1, 0, 2},
		},
		{
			name:   "move second before second",
			index:  1,
			before: 1,
			order:  []int{0, 1, 2},
		},
		{
			name:   "move second before third",
			index:  1,
			before: 2,
			order:  []int{0, 1, 2},
		},
		{
			name:   "move third before first",
			index:  2,
			before: 0,
			order:  []int{2, 0, 1},
		},
		{
			name:   "move third before second",
			index:  2,
			before: 1,
			order:  []int{0, 2, 1},
		},
		{
			name:   "move third before third",
			index:  2,
			before: 2,
			order:  []int{0, 1, 2},
		},
		{
			name:   "move non-existent before first",
			index:  -1,
			before: 0,
			order:  []int{0, 1, 2},
			err:    moodboard.ErrNoSuchItem,
		},
		{
			name:   "move first before non-existent",
			index:  0,
			before: -1,
			order:  []int{0, 1, 2},
			err:    moodboard.ErrNoSuchItem,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := memory.NewStore()

			var targetID string
			var beforeID string

			indexes := make(map[string]int)

			for i := 0; i < len(c.order); i++ {
				id, err := s.Create(bytes.NewReader(nil))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				if c.index == i {
					targetID = id
				}

				if c.before == i {
					beforeID = id
				}

				indexes[id] = i
			}

			switch err := s.MoveBefore(targetID, beforeID); {
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

			if len(all) != len(c.order) {
				verb := "items"

				if len(c.order) == 1 {
					verb = "item"
				}

				t.Fatalf("expected to get %d %s but got %d", len(c.order), verb, len(all))
			}

			for i, order := range c.order {
				id := all[i]
				index, ok := indexes[id]

				if !ok {
					t.Fatalf("no initial index for %q", id)
				}

				if index != order {
					t.Errorf("expected %q to be at index %d but got %d", id, order, i)
				}
			}
		})
	}
}

func TestStoreMoveAfter(t *testing.T) {
	cs := []struct {
		name  string
		index int
		after int
		order []int
		err   error
	}{
		{
			name:  "move first after first",
			index: 0,
			after: 0,
			order: []int{0, 1, 2},
		},
		{
			name:  "move first after second",
			index: 0,
			after: 1,
			order: []int{1, 0, 2},
		},
		{
			name:  "move first after third",
			index: 0,
			after: 2,
			order: []int{1, 2, 0},
		},
		{
			name:  "move second after first",
			index: 1,
			after: 0,
			order: []int{0, 1, 2},
		},
		{
			name:  "move second after second",
			index: 1,
			after: 1,
			order: []int{0, 1, 2},
		},
		{
			name:  "move second after third",
			index: 1,
			after: 2,
			order: []int{0, 2, 1},
		},
		{
			name:  "move third after first",
			index: 2,
			after: 0,
			order: []int{0, 2, 1},
		},
		{
			name:  "move third after second",
			index: 2,
			after: 1,
			order: []int{0, 1, 2},
		},
		{
			name:  "move third after third",
			index: 2,
			after: 2,
			order: []int{0, 1, 2},
		},
		{
			name:  "move non-existent after first",
			index: -1,
			after: 0,
			order: []int{0, 1, 2},
			err:   moodboard.ErrNoSuchItem,
		},
		{
			name:  "move first after non-existent",
			index: 0,
			after: -1,
			order: []int{0, 1, 2},
			err:   moodboard.ErrNoSuchItem,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			s := memory.NewStore()

			var targetID string
			var afterID string

			indexes := make(map[string]int)

			for i := 0; i < len(c.order); i++ {
				id, err := s.Create(bytes.NewReader(nil))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				if c.index == i {
					targetID = id
				}

				if c.after == i {
					afterID = id
				}

				indexes[id] = i
			}

			switch err := s.MoveAfter(targetID, afterID); {
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

			if len(all) != len(c.order) {
				verb := "items"

				if len(c.order) == 1 {
					verb = "item"
				}

				t.Fatalf("expected to get %d %s but got %d", len(c.order), verb, len(all))
			}

			for i, order := range c.order {
				id := all[i]
				index, ok := indexes[id]

				if !ok {
					t.Fatalf("no initial index for %q", id)
				}

				if index != order {
					t.Errorf("expected %q to be at index %d but got %d", id, order, i)
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
			s := memory.NewStore()
			items := make([]string, c.create)

			for i := 0; i < c.create; i++ {
				id, err := s.Create(bytes.NewReader(nil))

				if err != nil {
					t.Fatalf("failed to create item: %v", err)
				}

				items[i] = id
			}

			var id string

			if c.delete != -1 {
				id = items[c.delete]

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
				if all[i] != items[i] {
					t.Errorf("expected all[%d] to be %v but got %v", i, items[i], all[i])
				}
			}
		})
	}
}
