package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-info/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-info/pkg/info"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-info/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/xid"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultProtocol        = "http"
	defaultPort            = 8080
	defaultServerIp        = ""
	defaultServerPath      = "/"
	defaultSecondsToSleep  = 3
	secondsShutDownTimeout = 5 * time.Second  // maximum number of second to wait before closing server
	defaultReadTimeout     = 10 * time.Second // max time to read request from the client
	defaultWriteTimeout    = 10 * time.Second // max time to write response to the client
	defaultIdleTimeout     = 2 * time.Minute  // max time for connections using TCP Keep-Alive
	htmlHeaderStart        = `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css"/>`
	charsetUTF8            = "charset=UTF-8"
	MIMEAppJSON            = "application/json"
	MIMEHtml               = "text/html"
	MIMEAppJSONCharsetUTF8 = MIMEAppJSON + "; " + charsetUTF8
	HeaderContentType      = "Content-Type"
	httpErrMethodNotAllow  = "ERROR: Http method not allowed"
	initCallMsg            = "INITIAL CALL TO %s()\n"

	defaultNotFound            = "404 page not found"
	defaultNotFoundDescription = "🤔 ℍ𝕞𝕞... 𝕤𝕠𝕣𝕣𝕪 :【𝟜𝟘𝟜 : ℙ𝕒𝕘𝕖 ℕ𝕠𝕥 𝔽𝕠𝕦𝕟𝕕】🕳️ 🔥"
	formatTraceRequest         = "TRACE: [%s] %s  path:'%s', RemoteAddrIP: [%s]\n"
	formatErrRequest           = "ERROR: Http method not allowed [%s] %s  path:'%s', RemoteAddrIP: [%s]\n"
)

var rootPathGetCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_root_get_request_count", version.APP),
		Help: fmt.Sprintf("Number of GET request handled by %s default root handler", version.APP),
	},
)

var rootPathNotFoundCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_root_not_found_request_count", version.APP),
		Help: fmt.Sprintf("Number of page not found handled by %s default root handler", version.APP),
	},
)

type Middleware interface {
	// WrapHandler wraps the given HTTP handler for instrumentation.
	WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc
}

type middleware struct {
	buckets  []float64
	registry prometheus.Registerer
}

// WrapHandler wraps the given HTTP handler for instrumentation:
// It registers four metric collectors (if not already done) and reports HTTP
// metrics to the (newly or already) registered collectors.
// Each has a constant label named "handler" with the provided handlerName as
// value.
func (m *middleware) WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc {
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"handler": handlerName}, m.registry)

	requestsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)
	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: m.buckets,
		},
		[]string{"method", "code"},
	)
	requestSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",
		},
		[]string{"method", "code"},
	)
	responseSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",
		},
		[]string{"method", "code"},
	)

	// Wraps the provided http.Handler to observe the request result with the provided metrics.
	base := promhttp.InstrumentHandlerCounter(
		requestsTotal,
		promhttp.InstrumentHandlerDuration(
			requestDuration,
			promhttp.InstrumentHandlerRequestSize(
				requestSize,
				promhttp.InstrumentHandlerResponseSize(
					responseSize,
					handler,
				),
			),
		),
	)

	return base.ServeHTTP
}

// NewMiddleware returns a Middleware interface.
func NewMiddleware(registry prometheus.Registerer, buckets []float64) Middleware {
	if buckets == nil {
		buckets = prometheus.ExponentialBuckets(0.1, 1.5, 5)
	}

	return &middleware{
		buckets:  buckets,
		registry: registry,
	}
}

func getHtmlHeader(title string, description string) string {
	return fmt.Sprintf("%s<meta name=\"description\" content=\"%s\"><title>%s</title></head>", htmlHeaderStart, description, title)
}

func getHtmlPage(title string, description string) string {
	return getHtmlHeader(title, description) +
		fmt.Sprintf("\n<body><div class=\"container\"><h4>%s</h4></div></body></html>", title)
}

// WaitForHttpServer attempts to establish a TCP connection to listenAddress
// in a given amount of time. It returns upon a successful connection;
// otherwise exits with an error.
func WaitForHttpServer(listenAddress string, waitDuration time.Duration, numRetries int) {
	log.Printf("INFO: 'WaitForHttpServer Will wait for server to be up at %s for %v seconds, with %d retries'\n", listenAddress, waitDuration.Seconds(), numRetries)
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < numRetries; i++ {
		//conn, err := net.DialTimeout("tcp", listenAddress, dialTimeout)
		resp, err := httpClient.Get(listenAddress)
		if err != nil {
			fmt.Printf("\n[%d] Cannot make http get %s: %v\n", i, listenAddress, err)
			time.Sleep(waitDuration)
			continue
		}
		// All seems is good
		fmt.Printf("OK: Server responded after %d retries, with status code %d ", i, resp.StatusCode)
		return
	}
	log.Fatalf("Server %s not ready up after %d attempts", listenAddress, numRetries)
}

// waitForShutdownToExit will wait for interrupt signal SIGINT or SIGTERM and gracefully shutdown the server after secondsToWait seconds.
func waitForShutdownToExit(srv *http.Server, secondsToWait time.Duration) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	// wait for SIGINT (interrupt) 	: ctrl + C keypress, or in a shell : kill -SIGINT processId
	sig := <-interruptChan
	srv.ErrorLog.Printf("INFO: 'SIGINT %d interrupt signal received, about to shut down server after max %v seconds...'\n", sig, secondsToWait.Seconds())

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), secondsToWait)
	defer cancel()
	// gracefully shuts down the server without interrupting any active connections
	// as long as the actives connections last less than shutDownTimeout
	// https://pkg.go.dev/net/http#Server.Shutdown
	if err := srv.Shutdown(ctx); err != nil {
		srv.ErrorLog.Printf("💥💥 ERROR: 'Problem doing Shutdown %v'\n", err)
	}
	<-ctx.Done()
	srv.ErrorLog.Println("INFO: 'Server gracefully stopped, will exit'")
	os.Exit(0)
}

// GoHttpServer is a struct type to store information related to all handlers of web server
type GoHttpServer struct {
	listenAddress string
	// later we will store here the connection to database
	//DB  *db.Conn
	logger     *log.Logger
	router     *http.ServeMux
	registry   *prometheus.Registry
	startTime  time.Time
	httpServer http.Server
}

// NewGoHttpServer is a constructor that initializes the server mux (routes) and all fields of the  GoHttpServer type
func NewGoHttpServer(listenAddress string, logger *log.Logger) *GoHttpServer {
	myServerMux := http.NewServeMux()
	// Create non-global registry.
	registry := prometheus.NewRegistry()

	// Add go runtime metrics and process collectors.
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	registry.MustRegister(rootPathGetCounter)
	registry.MustRegister(rootPathNotFoundCounter)

	myServer := GoHttpServer{
		listenAddress: listenAddress,
		logger:        logger,
		router:        myServerMux,
		registry:      registry,
		startTime:     time.Now(),
		httpServer: http.Server{
			Addr:         listenAddress,       // configure the bind address
			Handler:      myServerMux,         // set the http mux
			ErrorLog:     logger,              // set the logger for the server
			ReadTimeout:  defaultReadTimeout,  // max time to read request from the client
			WriteTimeout: defaultWriteTimeout, // max time to write response to the client
			IdleTimeout:  defaultIdleTimeout,  // max time for connections using TCP Keep-Alive
		},
	}
	myServer.routes()

	return &myServer
}

// (*GoHttpServer) routes initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *GoHttpServer) routes() {
	// using new server Mux in Go 1.22 https://pkg.go.dev/net/http#ServeMux
	s.router.Handle("GET /{$}", NewMiddleware(
		s.registry, nil).
		WrapHandler("GET /$", s.getMyDefaultHandler()),
	)
	s.router.Handle("GET /time", s.getTimeHandler())
	s.router.Handle("GET /wait", s.getWaitHandler(defaultSecondsToSleep))
	s.router.Handle("GET /readiness", s.getReadinessHandler())
	s.router.Handle("GET /health", s.getHealthHandler())
	//expose the default prometheus metrics for Go applications
	s.router.Handle("GET /metrics", NewMiddleware(
		s.registry, nil).
		WrapHandler("GET /metrics", promhttp.HandlerFor(
			s.registry,
			promhttp.HandlerOpts{}),
		))

	s.router.Handle("GET /...", s.getHandlerNotFound())
}

// AddRoute   adds a handler for this web server
func (s *GoHttpServer) AddRoute(pathPattern string, handler http.Handler) {
	s.router.Handle(pathPattern, handler)
}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *GoHttpServer) StartServer() {

	// Starting the web server in his own goroutine
	go func() {
		s.logger.Printf("INFO: Starting http server listening at %s://%s/", defaultProtocol, s.listenAddress)
		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatalf("💥💥 ERROR: 'Could not listen on %q: %s'\n", s.listenAddress, err)
		}
	}()
	s.logger.Printf("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(&s.httpServer, secondsShutDownTimeout)

}

func (s *GoHttpServer) jsonResponse(w http.ResponseWriter, result interface{}) error {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Printf("ERROR: 'JSON marshal failed. Error: %v'", err)
		return err
	}
	var prettyOutput bytes.Buffer
	err = json.Indent(&prettyOutput, body, "", "  ")
	if err != nil {
		s.logger.Printf("ERROR: 'JSON Indent failed. Error: %v'", err)
		return err
	}
	w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(prettyOutput.Bytes())
	if err != nil {
		s.logger.Printf("ERROR: 'w.Write failed. Error: %v'", err)
		return err
	}
	return nil
}

//############# BEGIN HANDLERS

func (s *GoHttpServer) getReadinessHandler() http.HandlerFunc {
	handlerName := "getReadinessHandler"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
	}
}
func (s *GoHttpServer) getHealthHandler() http.HandlerFunc {
	handlerName := "getHealthHandler"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
	}
}
func (s *GoHttpServer) getMyDefaultHandler() http.HandlerFunc {
	handlerName := "getMyDefaultHandler"

	s.logger.Printf(initCallMsg, handlerName)

	data := info.CollectRuntimeInfo(s.logger)

	return func(w http.ResponseWriter, r *http.Request) {
		remoteIp := r.RemoteAddr // ip address of the original request or the last proxy
		requestedUrlPath := r.URL.Path
		guid := xid.New()
		s.logger.Printf("INFO: 'Request ID: %s'\n", guid.String())
		s.logger.Printf(formatTraceRequest, handlerName, r.Method, requestedUrlPath, remoteIp)
		query := r.URL.Query()
		nameValue := query.Get("name")
		if nameValue != "" {
			data.ParamName = nameValue
		}
		data.Hostname, _ = os.Hostname()
		data.RemoteAddr = remoteIp
		data.Headers = r.Header
		data.Uptime = fmt.Sprintf("%s", time.Since(s.startTime))
		uptimeOS, err := info.GetOsUptime()
		if err != nil {
			s.logger.Printf("💥💥 ERROR: 'GetOsUptime() returned an error : %+#v'", err)
		}
		data.UptimeOs = uptimeOS
		data.RequestId = guid.String()
		rootPathGetCounter.Inc()
		err = s.jsonResponse(w, data)
		if err != nil {
			s.logger.Printf("ERROR:  %v doing jsonResponse [%s] path:'%s', from IP: [%s]\n", err, handlerName, requestedUrlPath, remoteIp)
			return
		}
		s.logger.Printf("SUCCESS: [%s] path:'%s', from IP: [%s]\n", handlerName, requestedUrlPath, remoteIp)

	}
}
func (s *GoHttpServer) getHandlerNotFound() http.HandlerFunc {
	handlerName := "getHandlerNotFound"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		w.WriteHeader(http.StatusNotFound)
		rootPathNotFoundCounter.Inc()
		type NotFound struct {
			Status  int    `json:"status"`
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		data := &NotFound{
			Status:  http.StatusNotFound,
			Error:   defaultNotFound,
			Message: defaultNotFoundDescription,
		}
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			s.logger.Printf("💥💥 ERROR: [%s] Not Found was unable to Fprintf. path:'%s', from IP: [%s]\n", handlerName, r.URL.Path, r.RemoteAddr)
			http.Error(w, "Internal server error. myDefaultHandler was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}
func (s *GoHttpServer) getHandlerStaticPage(title string, descr string) http.HandlerFunc {
	handlerName := fmt.Sprintf("getHandlerStaticPage[%s]", title)
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set(HeaderContentType, MIMEHtml)
		w.WriteHeader(http.StatusOK)
		n, err := fmt.Fprintf(w, getHtmlPage(title, descr))
		if err != nil {
			s.logger.Printf("💥💥 ERROR: [%s]  was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d\n", handlerName, r.URL.Path, r.RemoteAddr, n)
			http.Error(w, "Internal server error. getHandlerStaticPage was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}
func (s *GoHttpServer) getTimeHandler() http.HandlerFunc {
	handlerName := "getTimeHandler"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		now := time.Now()
		w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintf(w, "{\"time\":\"%s\"}", now.Format(time.RFC3339))
		if err != nil {
			s.logger.Printf("Error doing fmt.Fprintf err: %s", err)
			return
		}
	}
}
func (s *GoHttpServer) getWaitHandler(secondsToSleep int) http.HandlerFunc {
	handlerName := "getWaitHandler"
	s.logger.Printf(initCallMsg, handlerName)
	durationOfSleep := time.Duration(secondsToSleep) * time.Second
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		if r.Method == http.MethodGet {
			w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
			time.Sleep(durationOfSleep) // simulate a delay to be ready
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintf(w, "{\"waited\":\"%v seconds\"}", secondsToSleep)
			if err != nil {
				s.logger.Printf("Error doing fmt.Fprintf err: %s", err)
				return
			}
		} else {
			s.logger.Printf(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, httpErrMethodNotAllow, http.StatusMethodNotAllowed)
		}
	}
}

// ############# END HANDLERS
func main() {
	listenAddr, err := config.GetPortFromEnv(defaultPort)
	if err != nil {
		log.Fatalf("💥💥 ERROR: 'calling GetPortFromEnv got error: %v'\n", err)
	}
	listenAddr = defaultServerIp + listenAddr
	l := log.New(os.Stdout, fmt.Sprintf("HTTP_SERVER_%s ", version.APP), log.Ldate|log.Ltime|log.Lshortfile)
	l.Printf("INFO: '🚀🚀 App %s version:%s  from %s'", version.APP, version.VERSION, version.REPOSITORY)
	l.Printf("INFO: 'Starting %s version:%s HTTP server on port %s'", version.APP, version.VERSION, listenAddr)
	server := NewGoHttpServer(listenAddr, l)
	// curl -vv  -X POST -H 'Content-Type: application/json'  http://localhost:8080/time   ==> 405 Method Not Allowed,
	// curl -vv  -X GET  -H 'Content-Type: application/json'  http://localhost:8080/time	==>200 OK , {"time":"2024-07-15T15:30:21+02:00"}
	server.AddRoute("GET /hello", server.getHandlerStaticPage("Hello", "Hello World!"))
	server.StartServer()
}
