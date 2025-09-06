# HTTP from TCP in Go

This project is a hands-on exploration of the HTTP protocol, built from the ground up using Go's networking capabilities. It demonstrates how to construct an HTTP server by directly working with TCP sockets, providing a deeper understanding of what happens under the hood of Go's built-in `net/http` package.

## Project Structure

The project is organized into the following directories:

-   `cmd`: Contains the main applications.
    -   `httpserver`: The main HTTP server application.
    -   `tcplistener`: A simple TCP listener that prints raw HTTP requests.
    -   `udpsender`: A simple UDP sender.
-   `internal`: Contains the core logic for the HTTP server.
    -   `headers`: Handles HTTP header parsing and manipulation.
    -   `request`: Responsible for parsing HTTP requests from a TCP connection.
    -   `response`: Provides tools for writing HTTP responses.
    -   `server`: The core TCP server that manages connections.
-   `assets`: Contains static assets, such as videos or images.

## How it Works

The server is built on the `net` package, which provides the foundation for TCP communication. Here's a high-level overview of the process:

1.  **Listen for TCP Connections:** The server starts by listening for incoming TCP connections on a specified port.
2.  **Accept Connections:** When a client connects, the server accepts the connection, creating a `net.Conn` object. This object represents the raw TCP connection and is used for reading and writing data.
3.  **Parse the HTTP Request:** The raw data from the TCP connection is read and parsed as an HTTP request. This involves:
    -   **Parsing the Request Line:** Identifying the HTTP method (e.g., `GET`, `POST`), the request target (e.g., `/`, `/about`), and the HTTP version.
    -   **Parsing Headers:** Reading the key-value pairs of the HTTP headers.
    -   **Reading the Body:** If the request has a body (e.g., in a `POST` request), it is read from the connection.
4.  **Handle the Request:** The parsed request is then passed to a handler function. This is where the application logic resides. The handler can be anything from serving a static file to proxying the request to another server.
5.  **Send the HTTP Response:** The handler generates an HTTP response, which is then written back to the client through the `net.Conn` object. This involves:
    -   **Writing the Status Line:** Sending the HTTP version, status code (e.g., `200 OK`), and reason phrase.
    -   **Writing Headers:** Sending the response headers.
    -   **Writing the Body:** Sending the response body.

## How to Run

### HTTP Server

To run the HTTP server, navigate to the `cmd/httpserver` directory and run the following command:

```bash
go run .
```

The server will start on port `42069`. You can then send requests to it using a tool like `curl` or your web browser.

### TCP Listener

The TCP listener is a useful tool for inspecting raw HTTP requests. To run it, navigate to the `cmd/tcplistener` directory and run:

```bash
go run .
```

You can then send a request to `localhost:42069` and see the raw request printed to the console.

### UDP Sender

The UDP sender is a simple application for sending UDP packets. To run it, navigate to the `cmd/udpsender` directory and run:

```bash
go run .
```

## Features

-   HTTP/1.1 compliant request parsing.
-   Support for various HTTP methods (`GET`, `POST`, etc.).
-   Static file serving.
-   Request proxying.
-   Chunked transfer encoding for responses.
-   Basic routing.

## Testing

To run the tests, execute the following command from the root of the project:

```bash
go test ./...
```

## Dependencies

This project uses the following external dependencies:

-   `github.com/stretchr/testify`: For assertions in tests.

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request.
