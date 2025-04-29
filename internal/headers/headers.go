package headers

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const crlf = "\r\n"

var validPattern = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+.^_|~`-]*$")

type Headers map[string]string

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h *Headers) Parse(data []byte) (int, bool, error) {

	if *h == nil {
		*h = make(Headers)
	}

	if len(data) == 0 {
		return 0, false, fmt.Errorf("no data to parse")
	}

	bytesConsumed := 0

	for {
		index := bytes.Index(data, []byte(crlf))
		if index == -1 {
			//log.Printf("No CRLF found in data")
			return bytesConsumed, false, nil
		}
		// if crlf is start of string, we're at the end of the header string
		if index == 0 {
			bytesConsumed += 2
			return bytesConsumed, true, nil
		}

		line := data[:index]
		parts := strings.SplitN(string(line), ":", 2)
		if len(parts) != 2 {
			return bytesConsumed, false, fmt.Errorf("invalid header format: missing or multiple colons")
		}
		if strings.HasSuffix(parts[0], " ") {
			return bytesConsumed, false, fmt.Errorf("invalid header format: whitespace before colon")
		}
		key := strings.TrimSpace(parts[0])
		if len(key) < 1 {
			return bytesConsumed, false, fmt.Errorf("invalid header format: field name must have at least one character")
		}
		if !validFieldName(key) {
			return bytesConsumed, false, fmt.Errorf("invalid header format: character in field name not permitted")
		}
		key = strings.ToLower(key)
		value := strings.TrimSpace(parts[1])

		_, ok := (*h)[key]
		if !ok {
			(*h)[key] = value
		} else {
			(*h)[key] = (*h)[key] + ", " + value
		}

		data = data[index+2:]
		bytesConsumed += index + 2
	}
}

func validFieldName(str string) bool {
	return validPattern.MatchString(str)
}

func (h Headers) Set(key, value string) Headers {
	newHeaders := h.Clone()
	newHeaders[key] = value
	return newHeaders
}

func (h Headers) Clone() Headers {
	newHeaders := make(Headers)
	for key, value := range h {
		newHeaders[key] = value
	}
	return newHeaders
}

func HttpCopy(httpH http.Header) Headers {
	h := Headers{}
	for k, v := range httpH {
		h.Set(k, strings.Join(v, ","))
	}
	return h
}
