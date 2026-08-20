# Overview

This is a web server written in Go.
I figure it's a great way to learn how to implement a web server from scratch in Go.

## Endpoints

### Admin

- `GET /admin/metrics` - Returns the number of hits to the server
- `POST /admin/reset` - Resets the database and the number of hits to the server if the PLATFORM environment variable is set to 'dev'

### API

- `POST /api/users` - Creates a new user
- `GET /api/chirps` - Returns all chirps
- `GET /api/chirps/{chirpID}` - Returns a specific chirp
- `POST /api/chirps` - Creates a new chirp