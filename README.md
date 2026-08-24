# Overview

This is a web server written in Go.
I figure it's a great way to learn how to implement a web server from scratch.
Obviosuly this is not a production-ready server and it's not intended to be used in a production environment.
Since it was my first time writing a web server, I'd like to note that this one was written following a guide.
Nevertheless I'm proud of the result and I'm sure it was a great learning experience.

## Features covered

- Authentication (JWT+Refresh Token)
- Authorization (via userID comparison)
- Metrics 
- Error handling
- Testing (unit tests)
- Database (PostgreSQL)
- Query Parameters

## Setup
The server has a config struct that loads the environment variables from a .env file in the root of the project.	
The .env file is not included in the repository (duh.).
```env
DB_URL=<db_url>
PLATFORM=<platform> ('dev' enables database reset functionality)
TOKEN_SECRET=<token_secret>
POLKA_KEY=<polka_key>
```

Make sure to have a PostgreSQL database running and to have the correct environment variables set.

To run the server, first install the package:
```bash
go install github.com/Drag0neUsz/Chirpy@latest
```

Then run the server:
```bash
Chirpy
```

The application will be available at http://localhost:8080.

## Endpoints

### Admin

- `GET /admin/metrics` - Returns the number of hits to the server
- `POST /admin/reset` - Resets the database and the number of hits to the server if the PLATFORM environment variable is set to 'dev'

### API

#### - `POST /api/users` - Creates a new user:

Request headers:
```
Content-Type: application/json
```
Request body:
```json
{
	"email": "test@example.com",
	"password": "password"
}
```
Response body:
```json
{
	"id": "123e4567-e89b-12d3-a456-426614174000",
	"created_at": "2021-01-01T00:00:00Z",
	"updated_at": "2021-01-01T00:00:00Z",
	"email": "test@example.com",
	"is_chirpy_red": false,
	"token": "xxx",
	"refresh_token": "xxx"
}
```

#### - `POST /api/login` - Logs in a user

Request headers:
```
Content-Type: application/json
```
Request body:
```json
{
	"email": "test@example.com",
	"password": "password"
}
```
Response body:
```json
{
	"id": "123e4567-e89b-12d3-a456-426614174000",
	"created_at": "2021-01-01T00:00:00Z",
	"updated_at": "2021-01-01T00:00:00Z",
	"email": "test@example.com",
	"is_chirpy_red": false,
	"token": "xxx",
	"refresh_token": "xxx"
}
```
#### - `PUT /api/users` - Updates a user

Request headers:
```
Authorization: Bearer <token>
Content-Type: application/json
```
Request body:
```json
{
	"email": "test@example.com",
	"password": "password"
}
```
Response body:
```json
{
	"id": "123e4567-e89b-12d3-a456-426614174000",
	"created_at": "2021-01-01T00:00:00Z",
	"updated_at": "2021-01-01T00:00:00Z",
	"email": "test@example.com",
	"is_chirpy_red": false
}
```
#### - `POST /api/refresh` - Refreshes a JWT token

Request headers:
```
Authorization: Bearer <refresh_token>
Content-Type: application/json
```
Response body:
```json
{
	"token": "xxx"
}
```
#### - `POST /api/revoke` - Revokes a refresh token

Request headers:
```
Authorization: Bearer <refresh_token>
Content-Type: application/json
```
#### - `GET /api/chirps` - Returns all chirps
Query parameters:
- `author_id=<author_id>` - Returns all chirps by a specific author
- `sort=<order>` - Returns all chirps in asc or desc order (default is ascending)
Response body:
```json
[
	{
		"id": "123e4567-e89b-12d3-a456-426614174000",
		"created_at": "2021-01-01T00:00:00Z",
		"updated_at": "2021-01-01T00:00:00Z",
		"body": "Hello, world!",
		"user_id": "123e4567-e89b-12d3-a456-426614174000"
	}
]
```
#### - `GET /api/chirps/{chirpID}` - Returns a specific chirp

Response body:
```json
{
	"id": "123e4567-e89b-12d3-a456-426614174000",
	"created_at": "2021-01-01T00:00:00Z",
	"updated_at": "2021-01-01T00:00:00Z",
	"body": "Hello, world!",
	"user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```
#### - `POST /api/chirps` - Creates a new chirp

Request headers:
```
Authorization: Bearer <token>
Content-Type: application/json
```
Request body:
```json
{
	"body": "Hello, world!"
}
```
Response body:
```json
{
	"id": "123e4567-e89b-12d3-a456-426614174000",
	"created_at": "2021-01-01T00:00:00Z",
	"updated_at": "2021-01-01T00:00:00Z",
	"body": "Hello, world!",
	"user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```
#### - `DELETE /api/chirps/{chirpID}` - Deletes a specific chirp

Request headers:
```
Authorization: Bearer <token>
```

### WebHooks

#### - `POST /api/webhooks/polka` - Handles Polka webhooks

Request body:
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "3311741c-680c-4546-99f3-fc9efac2036c"
  }
}
```