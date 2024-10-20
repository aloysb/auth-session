# Auth-Session

A very simple session-based authentication system.

## Usage

```bash
go run cmd/main.go
```

This will start a server on port 8080.

The server provides two routes:

- `/login` - creates a session 
- `/logout` - destroys the session
- `/authenticate` - validate the session


