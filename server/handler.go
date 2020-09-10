package moodboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// logger represents a simple logger.
type logger interface {
	Error(string)
}

// handler is a HTTP handler for moodboard requests.
type handler struct {
	logger logger
	store  *Store
}

// create handles inserting new moodboard entries.
func (h *handler) create(w http.ResponseWriter, r *http.Request) {
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
	if len(entry.URL) == 0 || entry.X < 0 || entry.X > 1 || entry.Y < 0 || entry.Y > 1 || entry.Width < 0 || entry.Width > 1 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err := h.store.Insert(entry)

	if errors.Is(err, ErrDuplicateURL) {
		w.WriteHeader(http.StatusConflict)
	} else if err != nil {
		// If we don't know how to handle this error then log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to insert entry: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// list handles listing moodboard entries.
func (h *handler) list(w http.ResponseWriter) {
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
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Accept", "application/json")

	// Make sure we have the right content type.
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)

		return
	}

	var entry struct {
		URL string `json:"url"`
	}

	// Try reading in the request.
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	// Make sure we have a URL in the request.
	if len(entry.URL) == 0 {
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	err := h.store.Delete(entry.URL)

	if errors.Is(err, ErrNoSuchEntry) {
		w.WriteHeader(http.StatusNotFound)
	} else if err != nil {
		// If we don't know how to handle this error then log it and return a generic error to the user.
		h.logger.Error(fmt.Sprintf("failed to delete entry: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.create(w, r)
	case http.MethodGet:
		h.list(w)
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
func NewHandler(l logger, s *Store) http.Handler {
	return &handler{logger: l, store: s}
}
