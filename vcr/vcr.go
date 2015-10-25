package vcr

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

func Client(vcrPath string) (*http.Client, io.Closer, error) {
	gz, e := os.Open(vcrPath)

	// Record VCR
	if e != nil {
		f, e := os.Create(vcrPath)
		if e != nil {
			return nil, f, e
		}
		gz := gzip.NewWriter(f)
		client := NewRecordingClient(gz)
		return &client, gz, nil
	}

	// Replay VCR
	f, e := gzip.NewReader(gz)
	if e != nil {
		return nil, f, e
	}
	client, e := NewReplayerClient(f)
	if e != nil {
		return nil, f, e
	}

	return &client, f, nil
}

func logRequest(w io.Writer, req *http.Request) error {
	b, err := httputil.DumpRequestOut(req, false)
	if err != nil {
		return fmt.Errorf("vcr: could not dump request: %s", err)
	}
	fmt.Fprint(w, string(b))
	return nil
}

func dumpResonse(w io.Writer, r *http.Response, body []byte) {
	// Status line
	text := r.Status
	protoMajor, protoMinor := strconv.Itoa(r.ProtoMajor), strconv.Itoa(r.ProtoMinor)
	statusCode := strconv.Itoa(r.StatusCode) + " "
	text = strings.TrimPrefix(text, statusCode)
	if _, err := io.WriteString(w, "HTTP/"+protoMajor+"."+protoMinor+" "+statusCode+text+"\r\n"); err != nil {
		panic(err)
	}

	r.Header.Write(w)

	if _, err := io.WriteString(w, "\r\n"); err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(body))
}

func logResponse(w io.Writer, res *http.Response, body bool) {
	var bodyBytes []byte
	var err error
	if body {
		defer res.Body.Close()
		bodyBytes, err = ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("could not record response: %s", err)
		}
		res.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	}
	dumpResonse(w, res, bodyBytes)
}

const recordSeparator string = "*************vcr*************\n"

func logSeparator(w io.Writer) {
	fmt.Fprintf(w, recordSeparator)
}

var defaultTransport = &http.Transport{}

type recorderTransport struct {
	mutex  sync.Mutex
	writer io.Writer
}

func (t *recorderTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	resp, err := defaultTransport.RoundTrip(r)
	if err != nil {
		panic(err)
	}

	logRequest(t.writer, r)
	logResponse(t.writer, resp, true)
	logSeparator(t.writer)

	return resp, err
}

func NewRecordingClient(w io.Writer) http.Client {
	client := http.Client{}
	client.Transport = &recorderTransport{writer: w}
	return client
}

func NewReplayerClient(r io.Reader) (http.Client, error) {
	client := http.Client{}
	transport, err := NewReplayerTransport(r)
	if err != nil {
		return client, err
	}
	client.Transport = transport
	return client, err
}

type replayerTransport struct {
	mutex     sync.Mutex
	requests  []*http.Request
	responses []*http.Response
}

func NewReplayerTransport(reader io.Reader) (*replayerTransport, error) {
	t := &replayerTransport{mutex: sync.Mutex{}}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	conversation, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("vcr: failed to read vcr file: %s", err)
	}
	r := bufio.NewReader(bytes.NewReader(conversation))

	i := 0
	for true {
		i++
		//		fmt.Printf("Reading Request %d\n", i)
		req, err := http.ReadRequest(r)
		if err == io.EOF {
			return t, nil
		} else if err != nil {
			fmt.Printf("error on request read: %s", err)
			return t, nil
		} else {
			t.requests = append(t.requests, req)
		}

		res, err := http.ReadResponse(r, req) // nil?
		if err == io.EOF {
			return t, nil
		} else if err != nil {
			fmt.Printf("error on response read: %s", err)
			return t, nil
		} else {
			res.Request.URL.Scheme = "http"
			res.Request.URL.Host = req.Host
			t.responses = append(t.responses, res)
		}

		bodyBytes := []byte{}
		for true {
			line, err := r.ReadBytes('\n')
			separatorReached := strings.HasSuffix(string(line), recordSeparator)
			if separatorReached {
				line = bytes.TrimSuffix(line, []byte(recordSeparator))
			}
			bodyBytes = append(bodyBytes, line...)

			if err == io.EOF || separatorReached {
				bodyReader := bytes.NewReader(bodyBytes)
				res.Body = ioutil.NopCloser(bodyReader)
				break
			} else if err == nil {
			} else {
				return t, err
			}
		}
	}

	return t, nil
}

func (t *replayerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	r, e := t.findResponse(req)
	if e != nil {
		return r, e
	}

	fmt.Fprintf(os.Stderr, "Replaying Request: %s %s://%s%s: %d\n",
		req.Method, req.URL.Scheme, req.URL.Host, req.URL.Path, r.StatusCode)

	r.Request.URL.Scheme = req.URL.Scheme
	return r, nil
}

func (t *replayerTransport) findResponse(req *http.Request) (*http.Response, error) {
	pattern := path.Join(req.URL.Host, req.URL.Path)
	for i, r := range t.requests {
		if r == nil {
			continue
		}

		hostPath := path.Join(r.URL.Host, r.URL.Path)
		if pattern == hostPath {
			t.requests[i] = nil // Mark as used
			return t.responses[i], nil
		}
	}
	return nil, errors.New("vcr: no matching request/response found")
}
