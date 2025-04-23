package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddress, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Unable to resolve UDP address: %v", err)
	}

	connection, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		log.Fatalf("Unable to open connection: %v", err)
	}
	defer connection.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading line: %v", err)
		}
		_, err = connection.Write([]byte(line))
		if err != nil {
			log.Printf("Error writing to connection: %v", err)
		}

	}

}
