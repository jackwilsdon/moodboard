# :sparkles: moodboard server :sparkles:

The moodboard server is a simple Go application which provides access to and management of moodboard entries via a RESTful API.

## Getting Started

To start the moodboard server, run the following commmmand:

```Text
$ go run cmd/moodboard/main.go data
```

This will start the moodboard server on port 3001, using the directory `data` as its data store.

### Executable Build

The moodboard server can also be built into a single executable, removing the runtime `go` dependency completely:

```Text
$ go build -o moodboard cmd/moodboard/main.go
```

```Text
$ ./moodboard data
```

## Stores

The moodboard server currently has two store implementations, [`file`](file) and [`memory`](memory).

To use the file-based store, pass a directory name on the command line:

```Text
$ ./moodboard data
```

To use the memory-based store, do not pass any arguments on the command line:

```Text
$ ./moodboard
```

**Note**: the memory-based store is not persisted across restarts, and as such should only be used for testing.
