package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/5tuartw/httpfromtcp/internal/request"
	"github.com/5tuartw/httpfromtcp/internal/response"
	"github.com/5tuartw/httpfromtcp/internal/server"
)

const port = 42069

func main() {

	handler := func(w *response.Writer, req *request.Request) {

		var statusCode response.StatusCode
		var htmlContent string

		if req.RequestLine.RequestTarget == "/yourproblem" {
			statusCode = response.BadRequest
			htmlContent = `<html>
							<head>
								<title>400 Bad Request</title>
							</head>
							<body>
								<h1>Bad Request</h1>
								<p>Your request honestly kinda sucked.</p>
							</body>
							</html>`
		} else if req.RequestLine.RequestTarget == "/myproblem" {
			statusCode = response.InternalServerError
			htmlContent = `<html>
							<head>
								<title>500 Internal Server Error</title>
							</head>
							<body>
								<h1>Internal Server Error</h1>
								<p>Okay, you know what? This one is on me.</p>
							</body>
							</html>`
		} else {
			statusCode = response.OK
			htmlContent = `<html>
							<head>
								<title>200 OK</title>
							</head>
							<body>
								<h1>Success!</h1>
								<p>Your request was an absolute banger.</p>
							</body>
							</html>`
		}

		bodyBytes := []byte(htmlContent)
		headers := response.GetDefaultHeaders(len(bodyBytes))
		headers = headers.Set("Content-Type", "text/html")

		w.WriteStatusLine(statusCode)
		w.WriteHeaders(headers)
		w.WriteBody(bodyBytes)

	}

	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
