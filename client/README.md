# :sparkles: moodboard client :sparkles:

The moodboard client is a simple React application which provides access to and management of moodboard items via a web based interface.

## Getting Started

To install dependencies for the moodboard client, run the following command:

```Text
$ npm install
```

The client can be started in development mode like so:

```Text
$ npm run start
```

This will start the moodboard client on port 3000, proxying all API requests to port 3001 (where the [server](../server) runs).

### Production Build

The moodboard client can be built into a production bundle, allowing it to be deployed to a static file server.

```Text
$ npm run build
```

This will output compiled assets to the `build` folder, where they can be deployed to a server.
