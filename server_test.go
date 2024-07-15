package main

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	DEBUG                           = false
	assertCorrectStatusCodeExpected = "expected status code should be returned"
	fmtErrNewRequest                = "### ERROR http.NewRequest %s on [%s] error is :%v\n"
	fmtTraceInfo                    = "### %s : %s on %s\n"
	fmtErr                          = "### GOT ERROR : %s\n%s"
	msgRespNotExpected              = "Response should contain what was expected."
)

type testStruct struct {
	name           string
	wantStatusCode int
	wantBody       string
	paramKeyValues map[string]string
	r              *http.Request
}

func TestErrorConfigError(t *testing.T) {
	err := ErrorConfig{
		err: errors.New("a brand new error test"),
		msg: "ERROR: This a test error.",
	}
	tests := []struct {
		name string
		e    ErrorConfig
		want string
	}{
		{
			name: "",
			e:    err,
			want: fmt.Sprintf("%s : %v", err.msg, err.err),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := err
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPortFromEnv(t *testing.T) {
	type args struct {
		defaultPort int
	}
	tests := []struct {
		name          string
		args          args
		envPORT       string
		want          string
		wantErr       bool
		wantErrPrefix string
	}{
		{
			name: "should return the default values when env variables are not set",
			args: args{
				defaultPort: defaultPort,
			},
			envPORT:       "",
			want:          ":8080",
			wantErr:       false,
			wantErrPrefix: "",
		},
		{
			name: "should return SERVERIP:PORT when env variables are set to valid values",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "3333",
			want:          ":3333",
			wantErr:       false,
			wantErrPrefix: "",
		},
		{
			name: "should return an empty string and report an error when PORT is not a number",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "aBigOne",
			want:          "",
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV PORT should contain a valid integer.",
		},
		{
			name: "should return an empty string and report an error when PORT is < 1",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "0",
			want:          "",
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
		},
		{
			name: "should return an empty string and report an error when PORT is > 65535",
			args: args{
				defaultPort: 8080,
			},
			envPORT:       "70000",
			want:          "",
			wantErr:       true,
			wantErrPrefix: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.envPORT) > 0 {
				err := os.Setenv("PORT", tt.envPORT)
				if err != nil {
					t.Errorf("Unable to set env variable PORT")
					return
				}
			}
			got, err := GetPortFromEnv(tt.args.defaultPort)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPortFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				// check that error contains the ERROR keyword
				if strings.HasPrefix(err.Error(), "ERROR:") == false {
					t.Errorf("GetPortFromEnv() error = %v, wantErrPrefix %v", err, tt.wantErrPrefix)
				}
			}
			if got != tt.want {
				t.Errorf("GetPortFromEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoHttpServerMyDefaultHandler(t *testing.T) {
	var l *log.Logger
	var nameParameter string
	listenAddr := fmt.Sprintf(":%d", defaultPort)
	if DEBUG {
		l = log.New(os.Stdout, fmt.Sprintf("HTTP_SERVER_%s ", APP), log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		l = log.New(io.Discard, APP, 0)
	}

	myServer := NewGoHttpServer(listenAddr, l)
	ts := httptest.NewServer(myServer.getMyDefaultHandler())
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
			name:           "1: Get on default Server Path should return a valid json containing param value",
			wantStatusCode: http.StatusOK,
			wantBody:       `"param_name": "╚»☯💥⚡✌ℂ𝔾𝕀𝕃✌⚡💥☯«╝"`,
			paramKeyValues: map[string]string{"name": "╚»☯💥⚡✌ℂ𝔾𝕀𝕃✌⚡💥☯«╝"},
			r:              newRequest(http.MethodGet, defaultServerPath, ""),
		},
		{
			name:           "2: Get on default Server Path should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, defaultServerPath, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set("Content-Type", "application/json")
			if len(tt.paramKeyValues) > 0 {
				parameters := tt.r.URL.Query()
				for paramName, paramValue := range tt.paramKeyValues {
					parameters.Add(paramName, paramValue)
					if paramName == "name" {
						nameParameter = paramValue
					}
				}
				tt.r.URL.RawQuery = parameters.Encode()
			}
			resp, err := http.DefaultClient.Do(tt.r)
			if DEBUG {
				fmt.Printf(fmtTraceInfo, tt.name, tt.r.Method, tt.r.URL)
			}
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)
			rInfo := &RuntimeInfo{}
			if DEBUG {
				fmt.Println("param name : % v", nameParameter)
				printWantedReceived(tt.wantBody, receivedJson)
			}
			if tt.wantStatusCode == http.StatusOK {
				err = json.Unmarshal(receivedJson, rInfo)
				assert.Nil(t, err, "the output should be a valid json")
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGoHttpServerReadinessHandler(t *testing.T) {
	myServer := NewGoHttpServer(fmt.Sprintf(":%d", defaultPort), log.New(io.Discard, APP, 0))
	ts := httptest.NewServer(myServer.getReadinessHandler())
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
			name:           "5: Get on readiness should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, "/readiness", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			resp, err := http.DefaultClient.Do(tt.r)
			if DEBUG {
				fmt.Printf(fmtTraceInfo, tt.name, tt.r.Method, tt.r.URL)
			}
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			if DEBUG {
				printWantedReceived(tt.wantBody, receivedJson)
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func printWantedReceived(wantBody string, receivedJson []byte) {
	fmt.Printf("WANTED   :%T - %#v\n", wantBody, wantBody)
	fmt.Printf("RECEIVED :%T - %#v\n", receivedJson, string(receivedJson))
}

func TestGoHttpServerHealthHandler(t *testing.T) {
	myServer := NewGoHttpServer(fmt.Sprintf(":%d", defaultPort), log.New(io.Discard, APP, 0))
	ts := httptest.NewServer(myServer.getHealthHandler())
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
			name:           "1: Get on health should return Http Status Ok",
			wantStatusCode: http.StatusOK,
			wantBody:       "",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, "/health", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			resp, err := http.DefaultClient.Do(tt.r)
			if DEBUG {
				fmt.Printf(fmtTraceInfo, tt.name, tt.r.Method, tt.r.URL)
			}
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			if DEBUG {
				printWantedReceived(tt.wantBody, receivedJson)
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGoHttpServerTimeHandler(t *testing.T) {
	myServer := NewGoHttpServer(fmt.Sprintf(":%d", defaultPort), log.New(os.Stdout, APP, log.Lshortfile))
	ts := httptest.NewServer(myServer.getTimeHandler())
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
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, "/time", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSON)
			resp, err := http.DefaultClient.Do(tt.r)
			if DEBUG {
				fmt.Printf(fmtTraceInfo, tt.name, tt.r.Method, tt.r.URL)
			}
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			if DEBUG {
				printWantedReceived(tt.wantBody, receivedJson)
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGoHttpServerWaitHandler(t *testing.T) {
	myServer := NewGoHttpServer(fmt.Sprintf(":%d", defaultPort), log.New(io.Discard, APP, 0))
	ts := httptest.NewServer(myServer.getWaitHandler(1))
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
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodGet, "/wait", ""),
		},
		{
			name:           "2: Post on /wait should return an http error method not allowed ",
			wantStatusCode: http.StatusMethodNotAllowed,
			wantBody:       "",
			paramKeyValues: make(map[string]string, 0),
			r:              newRequest(http.MethodPost, "/wait", `{"task":"test not allowed method "}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Header.Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			resp, err := http.DefaultClient.Do(tt.r)
			if DEBUG {
				fmt.Printf(fmtTraceInfo, tt.name, tt.r.Method, tt.r.URL)
			}
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)

			if DEBUG {
				printWantedReceived(tt.wantBody, receivedJson)
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestMainExecution(t *testing.T) {
	defaultPort := 9999
	listenAddr := fmt.Sprintf("%s://%s:%d", defaultProtocol, "127.0.0.1", defaultPort)
	err := os.Setenv("PORT", fmt.Sprintf("%d", defaultPort))
	if err != nil {
		t.Errorf("💥💥 ERROR: Unable to set env variable PORT")
		return
	}
	fmt.Printf("INFO: 'Will start HTTP server listening on port %s'\n", listenAddr)
	// starting main in his own go routine
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		main()
	}()
	WaitForHttpServer(listenAddr, 1*time.Second, 10)

	newRequest := func(method, url string, body string, useFormUrlencodedContentType bool) *http.Request {
		fmt.Printf("INFO: 🚀🚀'newRequest %s on %s ##BODY : %+v'\n", method, url, body)
		r, err := http.NewRequest(method, url, strings.NewReader(body))
		if err != nil {
			t.Fatalf(fmtErrNewRequest, method, url, err)
		}
		if method == http.MethodPost && useFormUrlencodedContentType {
			r.Header.Set(HeaderContentType, "application/x-www-form-urlencoded")
		} else {
			r.Header.Set(HeaderContentType, MIMEAppJSON)
		}
		return r
	}

	//resp, err := http.Get(listenAddr)
	//if err != nil {
	//	t.Fatalf("Cannot make http get: %v\n", err)
	//}
	//defer resp.Body.Close()
	//assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return an http status ok")
	//
	//receivedJson, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	t.Fatalf("Error reading response body: %v\n", err)
	//}
	//var decodedResponse OsInfo
	//err = json.Unmarshal(receivedJson, &decodedResponse)
	//assert.Nil(t, err, "the output should be a valid json")
	//if err != nil {
	//	t.Fatalf("Cannot decode response <%p> from server. Err: %v", receivedJson, err)
	//}
	//
	//// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
	//assert.Contains(t, string(receivedJson), fmt.Sprintf("\"appname\": \"%s\"", APP), "Response should contain the appname field.")
	//assert.Contains(t, string(receivedJson), "\"request_id\":", "Response should contain the request_id field.")

	type testStruct struct {
		name                         string
		contentType                  string
		wantStatusCode               int
		wantBody                     string
		paramKeyValues               map[string]string
		httpMethod                   string
		url                          string
		useFormUrlencodedContentType bool
		body                         string
	}

	tests := []testStruct{
		{
			name:                         "Get on default get handler should contain the appname field",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEAppJSON,
			wantBody:                     fmt.Sprintf("\"appname\": \"%s\"", APP),
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          "/",
			useFormUrlencodedContentType: false,
			body:                         "",
		},
		{
			name:                         "Post on default get handler should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  MIMEAppJSON,
			wantBody:                     "Method Not Allowed",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/",
			useFormUrlencodedContentType: true,
			body:                         `{"junk":"test with junk text"}`,
		},
		{
			name:                         "Get on nonexistent route should return an http error not found ",
			wantStatusCode:               http.StatusNotFound,
			contentType:                  MIMEAppJSON,
			wantBody:                     "page not found",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          "/aroutethatwillneverexisthere",
			useFormUrlencodedContentType: false,
			body:                         "",
		},
		{
			name:                         "Get on default Server Path should return a valid json containing param value",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEAppJSON,
			wantBody:                     `"param_name": "╚»☯💥⚡✌ℂ𝔾𝕀𝕃✌⚡💥☯«╝"`,
			paramKeyValues:               map[string]string{"name": "╚»☯💥⚡✌ℂ𝔾𝕀𝕃✌⚡💥☯«╝"},
			httpMethod:                   http.MethodGet,
			url:                          "/",
			useFormUrlencodedContentType: false,
			body:                         "",
		},
		{
			name:                         "/health Post should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  MIMEAppJSON,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/health",
			useFormUrlencodedContentType: false,
			body:                         `{"task":"test not allowed method "}`,
		},
		{
			name:                         "/readiness Post should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  MIMEAppJSON,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/readiness",
			useFormUrlencodedContentType: false,
			body:                         `{"task":"test not allowed method "}`,
		},
		{
			name:                         "/time Post should return an http error method not allowed ",
			wantStatusCode:               http.StatusMethodNotAllowed,
			contentType:                  MIMEAppJSON,
			wantBody:                     "",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodPost,
			url:                          "/time",
			useFormUrlencodedContentType: false,
			body:                         `{"task":"test not allowed method "}`,
		},
		{
			name:                         "/hello Get should return a welcome message",
			wantStatusCode:               http.StatusOK,
			contentType:                  MIMEHtml,
			wantBody:                     "Hello World!",
			paramKeyValues:               make(map[string]string, 0),
			httpMethod:                   http.MethodGet,
			url:                          "/hello",
			useFormUrlencodedContentType: false,
			body:                         ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare the request for this test case
			r := newRequest(tt.httpMethod, listenAddr+tt.url, tt.body, tt.useFormUrlencodedContentType)
			if len(tt.paramKeyValues) > 0 {
				parameters := r.URL.Query()
				for paramName, paramValue := range tt.paramKeyValues {
					parameters.Add(paramName, paramValue)
				}
				r.URL.RawQuery = parameters.Encode()
			}
			if DEBUG {
				fmt.Printf(fmtTraceInfo, tt.name, tt.httpMethod, tt.url)
			}
			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				fmt.Printf(fmtErr, err, resp.Body)
				t.Fatal(err)
			}
			defer resp.Body.Close()
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode, assertCorrectStatusCodeExpected)
			receivedJson, _ := io.ReadAll(resp.Body)
			rInfo := &RuntimeInfo{}
			if DEBUG {
				printWantedReceived(tt.wantBody, receivedJson)
			}
			if tt.wantStatusCode == http.StatusOK {
				if tt.contentType == MIMEAppJSON {
					err = json.Unmarshal(receivedJson, rInfo)
					assert.Nil(t, err, "the output should be a valid json")
				}
			}
			// check that receivedJson contains the specified tt.wantBody substring . https://pkg.go.dev/github.com/stretchr/testify/assert#Contains
			assert.Contains(t, string(receivedJson), tt.wantBody, msgRespNotExpected)
		})
	}
}

func TestGetKubernetesConnInfo(t *testing.T) {

	l := log.New(os.Stdout, APP, log.Lshortfile)

	type args struct {
		logger *log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *K8sInfo
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should return empty strings and an error when K8S_SERVICE_HOST is not set",
			args: args{logger: l},
			want: &K8sInfo{
				CurrentNamespace: "",
				Version:          "",
				Token:            "",
				CaCert:           "",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errConf := GetKubernetesConnInfo(tt.args.logger)
			if !tt.wantErr(t, errConf.err, fmt.Sprintf("GetKubernetesConnInfo() %s", tt.name)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetKubernetesConnInfo(%v)", tt.args.logger)
		})
	}
}

func TestGetJsonFromUrl(t *testing.T) {

	l := log.New(os.Stdout, APP, log.Lshortfile)
	expectedBody := `{"key": "value"}`
	// Create a mock server
	mockServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedBody))
	}))
	defer mockServer.Close()

	// Create a CA cert pool with the server's certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(mockServer.Certificate())

	// Create a mock server that will be closed to simulate a connection error
	mockServerClosed := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	// Create a CA cert pool with the server's certificate
	caCertPoolClosed := x509.NewCertPool()
	caCertPoolClosed.AddCert(mockServerClosed.Certificate())
	mockServerClosed.Close()

	// Create a mock server ReadError
	mockServerReadError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a partial write followed by a read error
		w.Header().Set("Content-Length", "1024")
		w.Write([]byte(expectedBody))

		// Close the connection prematurely to cause a read error
		conn, _, _ := w.(http.Hijacker).Hijack()
		conn.Close()
		// Close the connection immediately to cause a read error
		//w.(http.Flusher).Flush()
		//w.(http.CloseNotifier).CloseNotify()
	}))
	defer mockServerReadError.Close()

	type args struct {
		url           string
		bearerToken   string
		caCert        []byte
		allowInsecure bool
		logger        *log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should return an error when the url is not reachable",
			args: args{
				url:           "http://remotehostthatwillnotexist:9999",
				bearerToken:   "test-token",
				caCert:        nil,
				allowInsecure: false,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "should return status 200 ok when the url is reachable",
			args: args{
				url:           mockServer.URL,
				bearerToken:   "test-token",
				caCert:        mockServer.Certificate().Raw,
				allowInsecure: true,
				logger:        l,
			},
			want:    expectedBody,
			wantErr: assert.NoError,
		},
		{
			name: "should return an error when the url is reachable but the token is invalid",
			args: args{
				url:           mockServer.URL,
				bearerToken:   "",
				caCert:        mockServer.Certificate().Raw,
				allowInsecure: true,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "should return an error when the url is reachable but the connection is refused",
			args: args{
				url:           mockServerClosed.URL,
				bearerToken:   "test-token",
				caCert:        mockServerClosed.Certificate().Raw,
				allowInsecure: true,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
		{
			name: "should return an error when the url is reachable but response cannot be read",
			args: args{
				url:           mockServerReadError.URL,
				bearerToken:   "test-token",
				caCert:        nil,
				allowInsecure: true,
				logger:        l,
			},
			want:    "",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJsonFromUrl(tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.allowInsecure, tt.args.logger)
			if !tt.wantErr(t, err, fmt.Sprintf("GetJsonFromUrl(%v, %v, %v, %v)", tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.logger)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetJsonFromUrl(%v, %v, %v, %v)", tt.args.url, tt.args.bearerToken, tt.args.caCert, tt.args.logger)
		})
	}
}
