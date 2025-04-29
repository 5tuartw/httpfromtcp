package response

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/5tuartw/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	OK                  StatusCode = 200
	BadRequest          StatusCode = 400
	InternalServerError StatusCode = 500
)

func (s StatusCode) String() string {
	reason := http.StatusText(int(s))
	if reason == "" {
		return fmt.Sprintf("%d", int(s))
	}
	return fmt.Sprintf("%d %s", int(s), reason)
}

type WriterState int

const (
	WritingInitialised WriterState = iota
	WritingStatusDone
	WritingHeadersDone
	WritingBody
	WritingBodyDone
	WritingComplete
)

func (w WriterState) String() string {
	switch w {
	case WritingStatusDone:
		return "Status written, writing headers"
	case WritingHeadersDone:
		return "Headers written, writing body"
	case WritingBody:
		return "Writing body"
	case WritingBodyDone:
		return "Body written, writing trailers"
	case WritingComplete:
		return "Response writing complete"
	default:
		return fmt.Sprintf("Writing Status: %d", w)
	}
}

type Writer struct {
	State       WriterState
	IoWriter    io.Writer
	HasTrailers bool
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
	if w.State != WritingHeadersDone && w.State != WritingBody {
		return 0, fmt.Errorf("cannot write body while writer state is %s", w.State)
	}
	n, err := w.IoWriter.Write(p)
	if err != nil {
		return 0, err
	}
	//w.State = WritingComplete
	return n, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.State != WritingHeadersDone && w.State != WritingBody {
		return 0, fmt.Errorf("cannot write chunked body in current state: %s", w.State)
	}
	if w.State == WritingHeadersDone {
		w.State = WritingBody
	}

	dataLength := len(p)
	if dataLength == 0 {
		return 0, nil
	}

	hexLengthString := fmt.Sprintf("%x", dataLength)
	_, err := w.WriteBody([]byte(hexLengthString + "\r\n"))
	if err != nil {
		return 0, fmt.Errorf("error writing chunk size to body: %v", err)
	}
	all := append(p, []byte("\r\n")...)
	_, err = w.WriteBody(all)
	if err != nil {
		return 0, fmt.Errorf("error writing data chunk to body: %v", err)
	}
	return dataLength, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.State != WritingBody {
		return 0, fmt.Errorf("cannot write chunked body done in state %s", w.State)
	}
	var bytesWritten int
	var err error
	if w.HasTrailers {
		// Just write "0\r\n" when trailers will follow
		bytesWritten, err = fmt.Fprint(w.IoWriter, "0\r\n")
		if err != nil {
			return bytesWritten, err
		}
		w.State = WritingBodyDone
	} else {
		// Write "0\r\n\r\n" when no trailers
		bytesWritten, err = fmt.Fprint(w.IoWriter, "0\r\n\r\n")
		if err != nil {
			return bytesWritten, err
		}
		w.State = WritingComplete
	}

	return bytesWritten, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if !w.HasTrailers {
		return fmt.Errorf("no trailers declared in header")
	}
	if w.State != WritingBodyDone {
		return fmt.Errorf("cannot write trailers in state %s", w.State)
	}

	for key, value := range h {
		_, err := w.IoWriter.Write([]byte(key + ": " + value + crlf))
		if err != nil {
			return err
		}
	}
	_, err := w.IoWriter.Write([]byte(crlf))
	if err != nil {
		return err
	}
	w.State = WritingComplete
	return nil
}
