package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/5tuartw/httpfromtcp/internal/request"
	"github.com/5tuartw/httpfromtcp/internal/response"
)

type Server struct {
	Port       int
	Listener   net.Listener
	ServerOpen atomic.Bool
	Handler    Handler
}

func Serve(port int, h Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+fmt.Sprint(port))
	if err != nil {
		return nil, err
	}
	server := Server{
		Port:     port,
		Listener: listener,
		Handler:  h,
	}
	server.ServerOpen.Store(true)
	go server.listen()
	return &server, nil
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	s.ServerOpen.Store(false)
	return nil
}

func (s *Server) listen() {
	for s.ServerOpen.Load() {
		connection, err := s.Listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			} else {
				log.Printf("Error with connection: %v", err)
				continue
			}
		}
		go s.handle(connection)
	}
}

func (s *Server) handle(conn net.Conn) {
	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Could not parse request: %v", err)
		conn.Close()
		return
	}

	w := &response.Writer{
		State:    response.WritingInitialised,
		IoWriter: conn,
	}

	s.Handler(w, req)

	conn.Close()
}

type Handler func(w *response.Writer, req *request.Request)

