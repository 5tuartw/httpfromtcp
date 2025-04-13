package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	sourceFilePath := "messages.txt"
	file, err := os.Open(sourceFilePath)
	if err != nil {
		log.Fatalf("Error opening file '%s': %v", sourceFilePath, err)
	}
	defer file.Close()

	channel := getLinesChannel(file)

	for line := range channel {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
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
				/*currentLine = currentLine + parts[0]
				if len(parts) > 1 {
					stringChannel <- currentLine
					currentLine = parts[1]
				}*/
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
}
