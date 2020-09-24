package moodboard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// logger represents a simple logger.
type logger interface {
	Error(string)
}

// Handler is a HTTP handler for moodboard requests.
type Handler struct {
	logger logger
	store  Store
}

// validContentTypes is a list of allowed content types for uploaded images.
var validContentTypes = []string{
	"image/gif",
	"image/jpeg",
	"image/png",
}

// create handles reordering moodboard items.
func (h *Handler) move(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Accept", "application/json")

	// Make sure we have the right content type.
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	var target struct {
		Before string `json:"before"`
		After  string `json:"after"`
	}

	// Try reading in the request.
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	var err error

	// The ID of the item being moved comes after "/move/".
	id := r.URL.Path[6:]

	// We only want "before" or "after" - not both.
	if len(target.Before) > 0 && len(target.After) > 0 {
		w.WriteHeader(http.StatusBadRequest)

		return
	} else if len(target.Before) > 0 {
		err = h.store.MoveBefore(id, target.Before)
	} else if len(target.After) > 0 {
		err = h.store.MoveAfter(id, target.After)
	} else {
		// If we lack both "before" and "after" then it's a bad request.
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if errors.Is(err, ErrNoSuchItem) {
		w.WriteHeader(http.StatusNotFound)

		return
	} else if err != nil {
		// If we don't know how to handle this error then log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to move item: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// validateContentType checks the content type of the specified reader against validContentTypes.
//
// A new reader is returned which is prefixed with the result of any reads performed by this function.
func validateContentType(r io.Reader) (io.Reader, bool, error) {
	buf := make([]byte, 512)
	n, err := r.Read(buf)

	// If we got a non-EOF error, something else has gone wrong.
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, false, fmt.Errorf("failed to read header: %w", err)
	}

	// Only keep up to where we managed to read.
	buf = buf[:n]

	// Create a new reader which prefixes the reader we were given with the bytes we just read from it.
	r = io.MultiReader(bytes.NewReader(buf), r)

	contentType := http.DetectContentType(buf)

	// Check if the detected content type is in our valid type list.
	for _, validContentType := range validContentTypes {
		if contentType == validContentType {
			return r, true, nil
		}
	}

	return r, false, nil
}

// create handles inserting new moodboard items.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Accept", "multipart/form-data")

	mr, err := r.MultipartReader()

	if err != nil {
		// Make sure we have a multipart request.
		if errors.Is(err, http.ErrNotMultipart) || errors.Is(err, http.ErrMissingBoundary) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
		} else {
			// If we got some other error, it's probably the client's fault.
			w.WriteHeader(http.StatusBadRequest)
		}

		return
	}

	part, err := mr.NextPart()

	// If we got an error or the first part does not have the right name, the request is bad.
	if err != nil || part.FormName() != "file" {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Check the content type of the file being uploaded.
	partReader, isValid, err := validateContentType(part)

	if err != nil {
		// This error is unexpected - log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to detect content type: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// If the content type of the file isn't valid, return an error.
	if !isValid {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	id, err := h.store.Create(partReader)

	if err != nil {
		// This error is unexpected - log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to insert item: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(id)
}

// image handles getting images for moodboard items.
func (h *Handler) image(w http.ResponseWriter, r *http.Request) {
	// The ID of the image comes after "/image/".
	img, err := h.store.GetImage(r.URL.Path[7:])

	if errors.Is(err, ErrNoSuchItem) {
		w.WriteHeader(http.StatusNotFound)

		return
	} else if err != nil {
		// This error is unexpected - log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to get image: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Ask the client to cache the image.
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Pipe the image out to the response.
	_, _ = io.Copy(w, img)

	// Close the image if we can.
	if closer, ok := img.(io.ReadCloser); ok {
		_ = closer.Close()
	}
}

// list handles listing moodboard items.
func (h *Handler) list(w http.ResponseWriter) {
	es, err := h.store.All()

	// If we can't get a list of items then log the error and return a generic error to the client.
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed to list items: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// If we don't have any items then use a zero-length slice.
	//
	// This is needed to ensure that the JSON encoder does not return null instead of an empty array.
	if es == nil {
		es = make([]string, 0)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(es)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Accept", "application/json")

	// Make sure we have the right content type.
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	var item struct {
		ID string `json:"id"`
	}

	// Try reading in the request.
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Make sure we have an ID in the request.
	if len(item.ID) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err := h.store.Delete(item.ID)

	if errors.Is(err, ErrNoSuchItem) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		// If we don't know how to handle this error then log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to delete item: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if strings.HasPrefix(r.URL.Path, "/move/") {
			h.move(w, r)
		} else {
			h.create(w, r)
		}
	case http.MethodGet:
		if strings.HasPrefix(r.URL.Path, "/image/") {
			h.image(w, r)
		} else {
			h.list(w)
		}
	case http.MethodDelete:
		h.delete(w, r)
	default:
		w.Header().Add("Allow", "POST, GET, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// NewHandler creates a new moodboard HTTP handler.
func NewHandler(l logger, s Store) *Handler {
	return &Handler{logger: l, store: s}
}
