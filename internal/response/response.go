package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/5tuartw/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK StatusCode = iota
	BadRequest
	InternalServerError
)

func (s StatusCode) String() string {
	switch s {
	case OK:
		return "200 OK"
	case BadRequest:
		return "400 Bad Request"
	case InternalServerError:
		return "500 Internal Server Error"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

type WriterState int

const (
	WritingInitialised WriterState = iota
	WritingStatusDone
	WritingHeadersDone
	WritingComplete
)

func (w WriterState) String() string {
	switch w {
	case WritingStatusDone:
		return "Status written"
	case WritingHeadersDone:
		return "Headers written"
	case WritingComplete:
		return "Body written"
	default:
		return fmt.Sprintf("Writing Status: %d", w)
	}
}

type Writer struct {
	State    WriterState
	IoWriter io.Writer
}

const crlf = "\r\n"

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.Headers{}
	defaultHeaders["Content-Length"] = strconv.Itoa(contentLen)
	defaultHeaders["Connection"] = "close"
	defaultHeaders["Content-Type"] = "text/plain"
	return defaultHeaders
}

func (w *Writer) WriteStatusLine(s StatusCode) error {
	if w.State != WritingInitialised {
		return fmt.Errorf("cannot write status while writer state is %s", w.State)
	}
	_, err := w.IoWriter.Write([]byte("HTTP/1.1 " + s.String() + crlf))
	if err != nil {
		return err
	}
	w.State = WritingStatusDone
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != WritingStatusDone {
		return fmt.Errorf("cannot write headers while writer state is %s", w.State)
	}

	for key, value := range headers {
		_, err := w.IoWriter.Write([]byte(key + ": " + value + crlf))
		if err != nil {
			return err
		}
	}
	_, err := w.IoWriter.Write([]byte(crlf))
	if err != nil {
		return err
	}
	w.State = WritingHeadersDone
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != WritingHeadersDone {
		return 0, fmt.Errorf("cannot write body while writer state is %s", w.State)
	}
	n, err := w.IoWriter.Write(p)
	if err != nil {
		return 0, err
	}
	w.State = WritingComplete
	return n, nil
}

func (w * Writer) WriteChunkedBody(p []byte) (int, error) {

	return 0, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {

	return 0, nil
}