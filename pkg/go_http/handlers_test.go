package go_http

import (
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-info/pkg/tools"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-info/pkg/version"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	DEBUG                           = true
	assertCorrectStatusCodeExpected = "expected status code should be returned"
	fmtErr                          = "### GOT ERROR : %s\n%v"
	msgRespNotExpected              = "Response should contain what was expected."
)

var l *log.Logger

type testStruct struct {
	name           string
	wantStatusCode int
	wantBody       string
	paramKeyValues map[string]string
	r              *http.Request
}

func GetHttpTestRequest(t *testing.T, handler http.Handler, method, url string, body string) *http.Request {
	ts := httptest.NewServer(handler)
	defer ts.Close()
	r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
	if err != nil {
		t.Fatalf(fmtErrNewRequest, method, url, err)
	}
	return r
}

func executeTest(t *testing.T, tt testStruct, l *log.Logger) {
	t.Run(tt.name, func(t *testing.T) {
		tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		TraceRequest(tt.name, tt.r, l)
		resp, err := http.DefaultClient.Do(tt.r)
		if err != nil {
			fmt.Printf("Error doing http request for %s , Err: %v", tt.name, err)
			t.Fatal(err)
		}
		defer CloseBody(resp.Body, tt.name, l)
		assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
		receivedJson, _ := io.ReadAll(resp.Body)
		tools.PrintWantedReceived(tt.wantBody, receivedJson, l)
		// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
		assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
	})
}

func TestGoHttpServerHandlerNotFound(t *testing.T) {
	ts := httptest.NewServer(GetHandlerNotFound(l))
	defer ts.Close()

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}
	tests := []testStruct{
		{
			name:           "ARouteThatDoesNotExist GET should return Http Status 404 Not Found",
			wantStatusCode: http.StatusNotFound,
			wantBody:       "404 page not found",
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, "/ARouteThatDoesNotExist", ""),
		},
	}

	for _, tt := range tests {
		executeTest(t, tt, l)
	}
}

func TestGoHttpServerHandlerStaticPage(t *testing.T) {
	ts := httptest.NewServer(GetHandlerStaticPage("Title", "description", l))
	defer ts.Close()

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}
	tests := []testStruct{
		{
			name:           "GetHandlerStaticPage GET should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "<h4>Title</h4>",
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, "/hola", ""),
		},
	}
	for _, tt := range tests {
		executeTest(t, tt, l)
	}
}

func TestGoHttpServerHealthHandler(t *testing.T) {
	ts := httptest.NewServer(GetHealthHandler(l))
	defer ts.Close()

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}
	tests := []testStruct{
		{
			name:           "Get on health should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "",
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, "/health", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			resp, err := http.DefaultClient.Do(tt.r)
			TraceRequest(tt.name, tt.r, l)
			defer CloseBody(resp.Body, tt.name, l)
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			tools.PrintWantedReceived(tt.wantBody, receivedJson, l)
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGoHttpServerReadinessHandler(t *testing.T) {
	ts := httptest.NewServer(GetReadinessHandler(l))
	defer ts.Close()

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}

	tests := []testStruct{
		{
			name:           "readiness GET should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "",
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, "/readiness", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			resp, err := http.DefaultClient.Do(tt.r)
			TraceRequest(tt.name, tt.r, l)
			defer CloseBody(resp.Body, tt.name, l)
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)
			tools.PrintWantedReceived(tt.wantBody, receivedJson, l)
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGoHttpServerTimeHandler(t *testing.T) {
	ts := httptest.NewServer(GetTimeHandler(l))
	defer ts.Close()
	now := time.Now()
	expectedResult := fmt.Sprintf("{\"time\":\"%s\"}", now.Format(time.RFC3339))

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}

	tests := []testStruct{
		{
			name:           "1: Get on time should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       expectedResult,
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, "/time", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSON)
			resp, err := http.DefaultClient.Do(tt.r)
			TraceRequest(tt.name, tt.r, l)
			defer CloseBody(resp.Body, tt.name, l)
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			tools.PrintWantedReceived(tt.wantBody, receivedJson, l)
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGoHttpServerWaitHandler(t *testing.T) {
	ts := httptest.NewServer(GetWaitHandler(1, l))
	defer ts.Close()
	expectedResult := fmt.Sprintf("{\"waited\":\"%v seconds\"}", 1)

	newRequest := func(method, url string, body string) *http.Request {
		r, err := http.NewRequest(method, ts.URL+url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		return r
	}

	tests := []testStruct{
		{
			name:           "1: Get on /wait should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       expectedResult,
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodGet, "/wait", ""),
		},
		{
			name:           "2: Post on /wait should return an http error method not allowed ",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "",
			paramKeyValues: make(map[string]string),
			r:              newRequest(http.MethodPost, "/wait", `{"task":"test not allowed method "}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			resp, err := http.DefaultClient.Do(tt.r)
			TraceRequest(tt.name, tt.r, l)
			defer CloseBody(resp.Body, tt.name, l)
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			tools.PrintWantedReceived(tt.wantBody, receivedJson, l)
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func init() {
	if DEBUG {
		l = log.New(os.Stdout, fmt.Sprintf("testing_%s ", version.APP), log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		l = log.New(io.Discard, version.APP, 0)
	}
}
