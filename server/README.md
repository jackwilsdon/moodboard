# :sparkles: moodboard server :sparkles:

The moodboard server is a simple Go application implementing a RESTful API, backed by a JSON database.

## Getting Started

To start the moodboard server, run the following commmmand:

```Text
$ go run cmd/moodboard/main.go data.json
```

This will start the moodboard server on port 3001, using `data.json` as its data store.

### Executable Build

The moodboard server can also be built into a single executable, removing the runtime `go` dependency completely:

```Text
$ go build -o moodboard cmd/moodboard/main.go
```

```Text
$ ./moodboard data.json
```
