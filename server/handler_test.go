package moodboard_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackwilsdon/moodboard"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

type logger struct{}

func (logger) Error(string) {}

type store struct {
	t          *testing.T
	create     func(io.Reader) (string, error)
	all        func() ([]string, error)
	getImage   func(id string) (io.Reader, error)
	moveBefore func(id, beforeID string) error
	moveAfter  func(id, afterID string) error
	delete     func(id string) error
}

func (s store) Create(reader io.Reader) (string, error) {
	if s.create == nil {
		s.t.Fatalf("unexpected call to Create")
	}

	return s.create(reader)
}

func (s store) All() ([]string, error) {
	if s.all == nil {
		s.t.Fatalf("unexpected call to All")
	}

	return s.all()
}

func (s store) GetImage(id string) (io.Reader, error) {
	if s.getImage == nil {
		s.t.Fatalf("unexpected call to GetImage")
	}

	return s.getImage(id)
}

func (s store) MoveBefore(id, beforeID string) error {
	if s.moveBefore == nil {
		s.t.Fatalf("unexpected call to MoveBefore")
	}

	return s.moveBefore(id, beforeID)
}

func (s store) MoveAfter(id, afterID string) error {
	if s.moveAfter == nil {
		s.t.Fatalf("unexpected call to MoveAfter")
	}

	return s.moveAfter(id, afterID)
}

func (s store) Delete(id string) error {
	if s.delete == nil {
		s.t.Fatalf("unexpected call to Delete")
	}

	return s.delete(id)
}

var pngBytes []byte

func init() {
	buf := &bytes.Buffer{}

	if err := png.Encode(buf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		panic(fmt.Sprintf("failed to encode test PNG: %v", err))
	}

	pngBytes = buf.Bytes()
}

func TestMoveBefore(t *testing.T) {
	cs := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "no error returned from move function",
			statusCode: http.StatusOK,
		},
		{
			name:       "no such item error returned from move function",
			err:        moodboard.ErrNoSuchItem,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "unknown error returned from move function",
			err:        errors.New("something went wrong"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/move/id", bytes.NewBufferString(`{ "before": "beforeID" }`))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{
				t: t,
				moveBefore: func(id, beforeID string) error {
					if id != "id" {
						t.Errorf(`expected MoveBefore to be called with id "id" but got %q`, id)
					}

					if beforeID != "beforeID" {
						t.Errorf(`expected MoveBefore to be called with beforeID "beforeID" but got %q`, beforeID)
					}

					return c.err
				},
			}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			accepts := res.Header["Accept"]

			if len(accepts) == 1 {
				if accepts[0] != "application/json" {
					t.Errorf(`expected accept header to be "application/json" but got %q`, accepts[0])
				}
			} else {
				t.Errorf("expected 1 accept header but got %d", len(accepts))
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestMoveAfter(t *testing.T) {
	cs := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "no error returned from move function",
			statusCode: http.StatusOK,
		},
		{
			name:       "no such item error returned from move function",
			err:        moodboard.ErrNoSuchItem,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "unknown error returned from move function",
			err:        errors.New("something went wrong"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/move/id", bytes.NewBufferString(`{ "after": "afterID" }`))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{
				t: t,
				moveAfter: func(id, afterID string) error {
					if id != "id" {
						t.Errorf(`expected MoveAfter to be called with id "id" but got %q`, id)
					}

					if afterID != "afterID" {
						t.Errorf(`expected MoveAfter to be called with afterID "afterID" but got %q`, afterID)
					}

					return c.err
				},
			}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			accepts := res.Header["Accept"]

			if len(accepts) == 1 {
				if accepts[0] != "application/json" {
					t.Errorf(`expected accept header to be "application/json" but got %q`, accepts[0])
				}
			} else {
				t.Errorf("expected 1 accept header but got %d", len(accepts))
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestMoveInvalidContentType(t *testing.T) {
	cs := []string{"", "application/xml"}

	for _, c := range cs {
		name := c

		if name == "" {
			name = "empty"
		}

		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/move/id", bytes.NewBufferString("{ \"before\": \"beforeID\" }"))

			if c != "" {
				req.Header.Set("Content-Type", c)
			}

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{t: t}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != http.StatusUnsupportedMediaType {
				t.Errorf("expected status code to be %d but got %d", http.StatusUnsupportedMediaType, res.StatusCode)
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestMoveMalformedBody(t *testing.T) {
	cs := []struct {
		name string
		body string
	}{
		{
			name: "empty body",
		},
		{
			name: "malformed JSON",
			body: "{",
		},
		{
			name: "invalid type",
			body: "123",
		},
		{
			name: "no properties",
			body: "{}",
		},
		{
			name: "conflicting properties",
			body: `{ "before": "beforeID", "after": "afterID" }`,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/move/id", bytes.NewBufferString(c.body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{t: t}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status code to be %d but got %d", http.StatusBadRequest, res.StatusCode)
			}

			accepts := res.Header["Accept"]

			if len(accepts) == 1 {
				if accepts[0] != "application/json" {
					t.Errorf(`expected accept header to be "application/json" but got %q`, accepts[0])
				}
			} else {
				t.Errorf("expected 1 accept header but got %d", len(accepts))
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	cs := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "no error returned from create function",
			statusCode: http.StatusOK,
		},
		{
			name:       "error returned from move function",
			err:        errors.New("something went wrong"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			form := multipart.NewWriter(buf)
			file, _ := form.CreateFormFile("file", "example.png")
			_, _ = file.Write(pngBytes)
			_ = form.Close()

			req := httptest.NewRequest(http.MethodPost, "/", buf)
			req.Header.Set("Content-Type", form.FormDataContentType())

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{
				t: t,
				create: func(r io.Reader) (string, error) {
					buf, err := ioutil.ReadAll(r)

					if err != nil {
						t.Errorf("failed to read provided image: %v", err)
					} else if !bytes.Equal(pngBytes, buf) {
						t.Errorf("expected Create to be called with reader containing %q but got %q", pngBytes, buf)
					}

					return "id", c.err
				},
			}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			accepts := res.Header["Accept"]

			if len(accepts) == 1 {
				if accepts[0] != "multipart/form-data" {
					t.Errorf(`expected accept header to be "multipart/form-data" but got %q`, accepts[0])
				}
			} else {
				t.Errorf("expected 1 accept header but got %d", len(accepts))
			}

			if c.statusCode == http.StatusOK {
				contentType := res.Header["Content-Type"]

				if len(contentType) == 1 {
					if contentType[0] != "application/json; charset=utf-8" {
						t.Errorf(`expected content-type header to be "application/json; charset=utf-8" but got %q`, contentType[0])
					}
				} else {
					t.Errorf("expected 1 content-type header but got %d", len(contentType))
				}

				var id string

				if err := json.NewDecoder(res.Body).Decode(&id); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if id != "id" {
					t.Errorf(`expected response to be "id" but got %q`, id)
				}
			}
		})
	}
}

func TestCreateInvalidContentType(t *testing.T) {
	cs := []string{"", "application/xml", "multipart/form-data"}

	for _, c := range cs {
		name := c

		if name == "" {
			name = "empty"
		}

		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			if c != "" {
				req.Header.Set("Content-Type", c)
			}

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{t: t}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != http.StatusUnsupportedMediaType {
				t.Errorf("expected status code to be %d but got %d", http.StatusUnsupportedMediaType, res.StatusCode)
			}

			if count := len(res.Header["Content-Type"]); count > 0 {
				t.Errorf("expected no content-type header but got %d", count)
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestCreateMalformedBody(t *testing.T) {
	cs := []struct {
		name       string
		field      string
		file       []byte
		statusCode int
	}{
		{
			name:       "invalid file type",
			field:      "file",
			file:       []byte{0xCA, 0xFE, 0xBA, 0xBE},
			statusCode: http.StatusUnsupportedMediaType,
		},
		{
			name:       "wrong field name",
			field:      "wrong",
			file:       pngBytes,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			form := multipart.NewWriter(buf)
			file, _ := form.CreateFormFile(c.field, "example.png")
			_, _ = file.Write(c.file)
			_ = form.Close()

			req := httptest.NewRequest(http.MethodPost, "/", buf)
			req.Header.Set("Content-Type", form.FormDataContentType())

			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{t: t}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			accepts := res.Header["Accept"]

			if len(accepts) == 1 {
				if accepts[0] != "multipart/form-data" {
					t.Errorf(`expected accept header to be "multipart/form-data" but got %q`, accepts[0])
				}
			} else {
				t.Errorf("expected 1 accept header but got %d", len(accepts))
			}

			if count := len(res.Header["Content-Type"]); count > 0 {
				t.Errorf("expected no content-type header but got %d", count)
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestGetImage(t *testing.T) {
	cs := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "no error returned from image function",
			statusCode: http.StatusOK,
		},
		{
			name:       "no such item error returned from image function",
			err:        moodboard.ErrNoSuchItem,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "unknown error returned from image function",
			err:        errors.New("something went wrong"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/image/id", nil)
			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{
				t: t,
				getImage: func(id string) (io.Reader, error) {
					if id != "id" {
						t.Errorf(`expected GetImage to be called with id "id" but got %q`, id)
					}

					if c.err != nil {
						return nil, c.err
					}

					return bytes.NewReader(pngBytes), nil
				},
			}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			contentType := res.Header["Content-Type"]

			if c.statusCode == http.StatusOK {
				if len(contentType) == 1 {
					if contentType[0] != "image/png" {
						t.Errorf(`expected content-type header to be "image/png" but got %q`, contentType[0])
					}
				} else {
					t.Errorf("expected 1 content-type header but got %d", len(contentType))
				}

				body, _ := ioutil.ReadAll(res.Body)

				if !bytes.Equal(pngBytes, body) {
					t.Errorf("expected response to be %q but got %q", pngBytes, body)
				}
			} else if len(contentType) > 0 {
				t.Errorf("expected no content-type header but got %d", len(contentType))
			}
		})
	}
}

type reader struct {
	r      io.Reader
	closed bool
}

func (r reader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *reader) Close() error {
	r.closed = true

	if closer, ok := r.r.(io.Closer); ok {
		return closer.Close()
	}

	return nil
}

func TestGetImageClosesReader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/image/id", nil)
	w := httptest.NewRecorder()
	r := &reader{r: bytes.NewReader(pngBytes)}

	moodboard.NewHandler(logger{}, store{
		t: t,
		getImage: func(id string) (io.Reader, error) {
			if id != "id" {
				t.Errorf(`expected GetImage to be called with id "id" but got %q`, id)
			}

			return r, nil
		},
	}).ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status code to be %d but got %d", http.StatusOK, res.StatusCode)
	}

	contentType := res.Header["Content-Type"]

	if len(contentType) == 1 {
		if contentType[0] != "image/png" {
			t.Errorf(`expected content-type header to be "image/png" but got %q`, contentType[0])
		}
	} else {
		t.Errorf("expected 1 content-type header but got %d", len(contentType))
	}

	body, _ := ioutil.ReadAll(res.Body)

	if !bytes.Equal(pngBytes, body) {
		t.Errorf("expected response to be %q but got %q", pngBytes, body)
	}

	if !r.closed {
		t.Errorf("expected reader to be closed but got unclosed reader")
	}
}

func TestList(t *testing.T) {
	cs := []struct {
		name       string
		ids        []string
		err        error
		statusCode int
	}{
		{
			name:       "no IDs and no error returned from list function",
			ids:        nil,
			statusCode: http.StatusOK,
		},
		{
			name:       "no error returned from list function",
			ids:        []string{"first", "second", "third"},
			statusCode: http.StatusOK,
		},
		{
			name:       "error returned from list function",
			err:        errors.New("something went wrong"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{
				t: t,
				all: func() ([]string, error) {
					if c.err != nil {
						return nil, c.err
					}

					return c.ids, nil
				},
			}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			contentType := res.Header["Content-Type"]

			if c.statusCode == http.StatusOK {
				if len(contentType) == 1 {
					if contentType[0] != "application/json; charset=utf-8" {
						t.Errorf(`expected content-type header to be "application/json; charset=utf-8" but got %q`, contentType[0])
					}
				} else {
					t.Errorf("expected 1 content-type header but got %d", len(contentType))
				}

				var ids []string

				if err := json.NewDecoder(res.Body).Decode(&ids); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if len(ids) != len(c.ids) {
					t.Fatalf("expected 3 IDs but got %d", len(ids))
				}

				for i := range ids {
					if ids[i] != c.ids[i] {
						t.Errorf(`expected ids[%d] to be %q but got %q`, i, c.ids[i], ids[0])
					}
				}
			} else if len(contentType) > 0 {
				t.Errorf("expected no content-type header but got %d", len(contentType))
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cs := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "no error returned from delete function",
			statusCode: http.StatusOK,
		},
		{
			name:       "no such item error returned from delete function",
			err:        moodboard.ErrNoSuchItem,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "error returned from delete function",
			err:        errors.New("something went wrong"),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/id", nil)
			w := httptest.NewRecorder()

			moodboard.NewHandler(logger{}, store{
				t: t,
				delete: func(id string) error {
					if id != "id" {
						t.Errorf(`expected Delete to be called with id "id" but got %q`, id)
					}

					return c.err
				},
			}).ServeHTTP(w, req)

			res := w.Result()

			if res.StatusCode != c.statusCode {
				t.Errorf("expected status code to be %d but got %d", c.statusCode, res.StatusCode)
			}

			if count := len(res.Header["Content-Type"]); count > 0 {
				t.Errorf("expected no content-type header but got %d", count)
			}

			body, _ := ioutil.ReadAll(res.Body)

			if len(body) > 0 {
				t.Errorf("expected empty body but got %q", body)
			}
		})
	}
}

func TestInvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	w := httptest.NewRecorder()

	moodboard.NewHandler(logger{}, store{t: t}).ServeHTTP(w, req)

	res := w.Result()

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status code to be %d but got %d", http.StatusMethodNotAllowed, res.StatusCode)
	}

	allows := res.Header["Allow"]

	if len(allows) == 1 {
		if allows[0] != "POST, GET, DELETE" {
			t.Errorf(`expected allow header to be "POST, GET, DELETE" but got %q`, allows[0])
		}
	} else {
		t.Errorf("expected 1 allow header but got %d", len(allows))
	}

	if count := len(res.Header["Content-Type"]); count > 0 {
		t.Errorf("expected no content-type header but got %d", count)
	}

	body, _ := ioutil.ReadAll(res.Body)

	if len(body) > 0 {
		t.Errorf("expected empty body but got %q", body)
	}
}
