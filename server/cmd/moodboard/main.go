package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackwilsdon/moodboard"
)

type logger struct{}

func (logger) Error(msg string) {
	log.Print(msg)
}

func main() {
	// Ensure a store is provided.
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s data.json\n", os.Args[0])
		os.Exit(1)
	}

	// Create a new store with the provided path.
	s := moodboard.NewStore(os.Args[1])

	// Handle requests to the root with the moodboard handler.
	http.Handle("/", moodboard.NewHandler(logger{}, s))

	log.Print("starting on http://localhost:3001...")

	// Start the server on port 3001.
	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatal(err)
	}
}
