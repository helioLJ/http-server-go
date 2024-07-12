# Go HTTP Server

This is a simple HTTP server written in Go that handles various types of requests and supports concurrent connections.

## Features

- Handles GET and POST requests
- Supports concurrent connections
- Implements several endpoints:
- Root endpoint (`/`)
- Echo endpoint (`/echo/<message>`)
- User-Agent endpoint (`/user-agent`)
- File handling endpoints (`/files/<filename>`)
- Supports gzip compression for responses
- Logs incoming requests and their statuses

## Prerequisites

- Go 1.16 or higher

## Installation

1. Clone this repository:
git clone https://github.com/helioLJ/http-server-go.git
cd http-server-go


2. Build the server:
./start_server.sh


## Usage

### Starting the Server

Run the server using the start script:

./start_server.sh [--directory <path>]


The `--directory` flag is optional and specifies the directory for file operations. If not provided, the current directory will be used.

### Endpoints

1. **Root Endpoint (`/`)**
   - Returns a 200 OK status with an empty body.

2. **Echo Endpoint (`/echo/<message>`)**
   - Returns the `<message>` in the response body.

3. **User-Agent Endpoint (`/user-agent`)**
   - Returns the User-Agent header from the request.

4. **File Handling Endpoints (`/files/<filename>`)**
   - GET: Retrieves the content of the specified file.
   - POST: Creates a new file with the specified name and content.

### Concurrent Connections

The server supports handling multiple connections concurrently. You can test this using the provided `concurrent_connections.sh` script:

./concurrent_connections.sh


This script sends three concurrent requests to the server.

## Configuration

The server listens on port 4221 by default. To change this, modify the `net.Listen()` call in the `main()` function of `server.go`.

## Logging

The server logs each request with the following information:
- Timestamp
- Remote address
- HTTP method
- Requested path
- Response status

Logs are printed to the console.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
