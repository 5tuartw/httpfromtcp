package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/5tuartw/httpfromtcp/internal/headers"
	"github.com/5tuartw/httpfromtcp/internal/request"
	"github.com/5tuartw/httpfromtcp/internal/response"
	"github.com/5tuartw/httpfromtcp/internal/server"
)

const port = 42069

func main() {

	handler := func(w *response.Writer, req *request.Request) {

		var statusCode response.StatusCode
		var htmlContent string
		var target = req.RequestLine.RequestTarget

		if strings.HasPrefix(target, "/httpbin") {
			target = "https://httpbin.org" + strings.TrimPrefix(target, "/httpbin")

			resp, err := http.Get(target)
			if err != nil {
				log.Printf("error forwarding request to %s: %v", target, err)
				w.WriteChunkedBodyDone()
				return
			}
			defer resp.Body.Close()

			responseStatus := response.StatusCode(resp.StatusCode)
			w.WriteStatusLine(responseStatus)

			responseHeaders := headers.HttpCopy(resp.Header)

			delete(responseHeaders, "Content-Length")
			responseHeaders["Transfer-Encoding"] = "chunked"
			responseHeaders["Trailer"] = "X-Content-SHA256, X-Content-Length"
			w.HasTrailers = true

			w.WriteHeaders(responseHeaders)

			var fullBody []byte
			buffer := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buffer)
				if n > 0 {
					_, errWrite := w.WriteChunkedBody(buffer[:n])
					if errWrite != nil {
						log.Printf("error writing chunked body: %v", errWrite)
						return
					}
					fullBody = append(fullBody, buffer[:n]...)
				}
				if err != nil {
					if err != io.EOF {
						log.Printf("error reading from target %s: %v", target, err)
					}
					break
				}
				if n == 0 {
					break
				}
			}

			_, err = w.WriteChunkedBodyDone()
			if err != nil {
				log.Printf("error writing chunked body done: %v", err)
				return
			}

			hash := sha256.Sum256(fullBody)
			hashHex := hex.EncodeToString(hash[:])
			contentLength := len(fullBody)
			trailers := headers.Headers{
				"X-Content-SHA256": hashHex,
				"X-Content-Length": strconv.Itoa(contentLength),
			}
			w.WriteTrailers(trailers)
			

			return
		} else if target == "/video" {
			responseHeaders := response.GetDefaultHeaders(0)

			videoData, err := os.ReadFile("assets/vim.mp4")
			if err != nil {
				statusCode = response.InternalServerError
				w.WriteStatusLine(statusCode)
				w.WriteHeaders(responseHeaders)
				w.WriteBody([]byte("Internal Server Error: Could not load video"))
			} else {
				statusCode = response.OK
				responseHeaders["Content-Length"] = fmt.Sprintf("%d", len(videoData))
				responseHeaders["Content-Type"] = "video/mp4"
				w.WriteStatusLine(statusCode)
				w.WriteHeaders(responseHeaders)
				w.WriteBody(videoData)


			}

			} else {


				if target == "/yourproblem" {
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
				} else if target == "/myproblem" {
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
				responseHeaders := response.GetDefaultHeaders(len(bodyBytes))
				responseHeaders = responseHeaders.Set("Content-Type", "text/html")

				w.WriteStatusLine(statusCode)
				w.WriteHeaders(responseHeaders)
				w.WriteBody(bodyBytes)
		}

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
