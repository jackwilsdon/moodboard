package main

import (
	"fmt"
	"github.com/jackwilsdon/moodboard/file"
	"github.com/jackwilsdon/moodboard/memory"
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
	var s moodboard.Store

	// Create the right type of store based on the number of arguments we were given.
	if len(os.Args) == 1 {
		s = memory.NewStore(nil)

		log.Print("using in-memory store")
	} else if len(os.Args) == 2 {
		s = file.NewStore(os.Args[1])

		log.Printf("using file-based store %q", os.Args[1])
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %s [data.json]\n", os.Args[0])
		os.Exit(1)
	}

	// Handle requests to the root with the moodboard handler.
	http.Handle("/", moodboard.NewHandler(logger{}, s))

	log.Print("starting on http://localhost:3001...")

	// Start the server on port 3001.
	if err := http.ListenAndServe(":3001", nil); err != nil {
		log.Fatal(err)
	}
}
