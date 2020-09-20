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

// create handles inserting new moodboard entries.
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

	entry, err := h.store.Create(partReader)

	if err != nil {
		// This error is unexpected - log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to insert entry: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(entry.ID)
}

// image handles getting images for moodboard entries.
func (h *Handler) image(w http.ResponseWriter, r *http.Request) {
	// The ID of the image comes after "/image/".
	img, err := h.store.GetImage(r.URL.Path[7:])

	if errors.Is(err, ErrNoSuchEntry) {
		w.WriteHeader(http.StatusNotFound)

		return
	} else if err != nil {
		// This error is unexpected - log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to get image: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Pipe the image out to the response.
	_, _ = io.Copy(w, img)

	// Close the image if we can.
	if closer, ok := img.(io.ReadCloser); ok {
		_ = closer.Close()
	}
}

// list handles listing moodboard entries.
func (h *Handler) list(w http.ResponseWriter) {
	es, err := h.store.All()

	// If we can't get a list of entries then log the error and return a generic error to the client.
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed to list entries: %v", err))
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// If we don't have any entries then use a zero-length slice.
	//
	// This is needed to ensure that the JSON encoder does not return null instead of an empty array.
	if es == nil {
		es = make([]Entry, 0)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(es)
}

// update handles updating existing moodboard entries.
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Accept", "application/json")

	// Make sure we have the right content type.
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	var entry Entry

	// Try reading in the request.
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Make sure the request is valid.
	if entry.X < 0 || entry.X > 1 || entry.Y < 0 || entry.Y > 1 || entry.Width < 0 || entry.Width > 1 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err := h.store.Update(entry)

	if errors.Is(err, ErrNoSuchEntry) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		// If we don't know how to handle this error then log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to update entry: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Accept", "application/json")

	// Make sure we have the right content type.
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	var entry struct {
		ID string `json:"id"`
	}

	// Try reading in the request.
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Make sure we have an ID in the request.
	if len(entry.ID) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err := h.store.Delete(entry.ID)

	if errors.Is(err, ErrNoSuchEntry) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		// If we don't know how to handle this error then log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to delete entry: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.create(w, r)
	case http.MethodGet:
		if strings.HasPrefix(r.URL.Path, "/image/") {
			h.image(w, r)
		} else {
			h.list(w)
		}
	case http.MethodPut:
		h.update(w, r)
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
