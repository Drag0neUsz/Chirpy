# Overview

This is a web server written in Go.
I figure it's a great way to learn how to implement a web server from scratch in Go.

## Features covered

- Authentication (JWT+Refresh Token)
- Authorization (via userID comparison)
- Metrics 
- Error handling
- Testing (unit tests)


## Endpoints

### Admin

- `GET /admin/metrics` - Returns the number of hits to the server
- `POST /admin/reset` - Resets the database and the number of hits to the server if the PLATFORM environment variable is set to 'dev'

### API

- `POST /api/users` - Creates a new user
- `POST /api/login` - Logs in a user
- `PUT /api/users` - Updates a user
- `POST /api/refresh` - Refreshes a JWT token
- `POST /api/revoke` - Revokes a refresh token
- `GET /api/chirps` - Returns all chirps
- `GET /api/chirps/{chirpID}` - Returns a specific chirp
- `POST /api/chirps` - Creates a new chirp
- `DELETE /api/chirps/{chirpID}` - Deletes a specific chirp