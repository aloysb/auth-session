# Auth-Session

A very simple session-based authentication system.

## Usage
```bash
go run cmd/main.go
```

This will start a server on port 8080 by default.

The server provides two routes:

- `/login` - creates a session 
- `/logout` - destroys the session
- `/authenticate` - validate the session

See the open API spec for more information in the `/api/openapi.yaml` file.

## Configuration

The configuration is done via environment variables:
- `DATABASE_URL` - The database connection string. Currently only SQLite is supported
- `DATABASE_TYPE` - The database type, defaults to `sqlite`. Currently only `sqlite` is supported

## Limitations

- The authentication has to be handled by the client at this stage, this library is purely a session manager.
- There is only an SQLite adapter at this point, and no ability to connect to external databases.
- The listening port is hardcoded to 8080.

## Roadmap
- Add support for PostgreSQL
- Add dockerfile and docker-compose example
