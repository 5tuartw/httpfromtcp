package main

import (
	"fmt"
	"log"
	"net"

	"github.com/5tuartw/httpfromtcp/internal/request"
)

func main() {

	port := ":42069"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error opening listener on port %s: %v", port, err)
		return
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatalf("Could not accept connection: %v", err)
			return
		}

		log.Printf("Connection has been accepted from %s.", connection.RemoteAddr())

		request, err := request.RequestFromReader(connection)
		if err != nil {
			log.Fatalf("Unable to receive request")
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range request.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(request.Body))
		fmt.Println("...")

		log.Printf("Connection from %s has been closed.", connection.RemoteAddr())
	}
}

/*func getLinesChannel(f io.ReadCloser) <-chan string {
	stringChannel := make(chan string)

	go func() {
		currentLine := ""
		bufferSize := 8
		buffer := make([]byte, bufferSize)
		for {
			n, err := f.Read(buffer)
			if n > 0 {
				parts := strings.Split(string(buffer[:n]), "\n")
				parts[0] = currentLine + parts[0]

				for i := 0; i < len(parts)-1; i++ {
					stringChannel <- parts[i]
				}
				currentLine = parts[len(parts)-1]
			}
			if err != nil {
				if err == io.EOF {
					if currentLine != "" {
						stringChannel <- currentLine
					}
					break
				}
				log.Fatalf("Error reading file: %v", err)
			}
		}
		close(stringChannel)
		f.Close()
	}()
	return stringChannel
}*/
