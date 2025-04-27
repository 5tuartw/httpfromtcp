package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/5tuartw/httpfromtcp/internal/headers"
)

type Status int

const (
	requestStateInitialised Status = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

func (s Status) String() string {
	switch s {
	case requestStateInitialised:
		return "initialised"
	case requestStateParsingHeaders:
		return "parsing headers"
	case requestStateParsingBody:
		return "parsing body"
	case requestStateDone:
		return "done"
	default:
		return fmt.Sprintf("Status(%d)", s)
	}
}

type Request struct {
	RequestLine RequestLine
	ParserState Status
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	request := &Request{
		ParserState: requestStateInitialised,
		Headers:     make(headers.Headers),
	}

	for request.ParserState != requestStateDone {
		// If buffer is full, grow it
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading from reader: %v", err)
		}
		readToIndex += n

		bytesConsumed, parseErr := request.Parse(buf[:readToIndex])
		if parseErr != nil {
			return nil, parseErr
		}

		// remove consumed data from buffer
		if bytesConsumed > 0 {
			copy(buf, buf[bytesConsumed:readToIndex])
			readToIndex -= bytesConsumed
		}

		if err == io.EOF {
			break
		}
	}

	if request.ParserState != requestStateDone {
		return nil, fmt.Errorf("incomplete request: reached EOF before finding end of request line")
	}

	validMethod := validateMethod(request.RequestLine.Method)
	if !validMethod {
		return nil, fmt.Errorf("invalid method: %s ", request.RequestLine.Method)
	}

	if !(request.RequestLine.HttpVersion == "1.1") {
		return nil, errors.New("only HTTP/1.1 is supported")
	}

	return request, nil
}

func (r *Request) Parse(data []byte) (int, error) {
	totalBytesParsed := 0

	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	if r.ParserState == requestStateInitialised {
		requestLine, numBytes, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if numBytes == 0 {
			return 0, nil
		} else {
			r.RequestLine = *requestLine
			r.ParserState = requestStateParsingHeaders
			return numBytes, nil
		}
	} else if r.ParserState == requestStateParsingHeaders {
		if len(data) == 0 {
			return 0, nil
		}
		bytesRead, isDone, err := (&r.Headers).Parse(data)
		if err != nil {
			return bytesRead, err
		}
		if isDone {
			r.ParserState = requestStateParsingBody
		}
		return bytesRead, nil
	} else if r.ParserState == requestStateParsingBody {
		length := r.Headers.Get("Content-Length")
		if length == "" {
			r.ParserState = requestStateDone
			return 0, nil
		}

		contentLength, err := strconv.Atoi(length)
		if err != nil {
			return 0, fmt.Errorf("invalid Content-Length: %v", err)
		}

		// Append the current chunk to the body
		r.Body = append(r.Body, data...)

		// Check if we've read enough bytes
		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("body length exceeds content length")
		} else if len(r.Body) == contentLength {
			r.ParserState = requestStateDone
		}

		// Return that we've consumed all the data in this chunk
		return len(data), nil
	} else if r.ParserState == requestStateDone {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	} else {
		return 0, fmt.Errorf("error: unknown state")
	}
}

func parseRequestLine(rLine []byte) (*RequestLine, int, error) {
	idx := bytes.Index(rLine, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(rLine[:idx])

	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + len(crlf), nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	rLineParts := strings.Split(str, " ")

	if len(rLineParts) != 3 {
		err := errors.New("request line has too few arguements: " + str)
		return nil, err
	}

	httpFull := rLineParts[2]
	version := strings.TrimPrefix(httpFull, "HTTP/")

	requestLine := RequestLine{
		HttpVersion:   version,
		RequestTarget: rLineParts[1],
		Method:        rLineParts[0],
	}

	return &requestLine, nil
}

func validateMethod(method string) bool {
	valid := true

	for _, char := range method {
		if !unicode.IsUpper(char) {
			valid = false
			break
		}
	}

	return valid
}
