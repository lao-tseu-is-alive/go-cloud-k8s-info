package go_http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//############# BEGIN HANDLERS

func (s *GoHttpServer) GetReadinessHandler() http.HandlerFunc {
	handlerName := "GetReadinessHandler"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, s.logger)
		w.WriteHeader(http.StatusOK)
	}
}
func (s *GoHttpServer) GetHealthHandler() http.HandlerFunc {
	handlerName := "GetHealthHandler"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, s.logger)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *GoHttpServer) GetHandlerNotFound() http.HandlerFunc {
	handlerName := "GetHandlerNotFound"
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
func (s *GoHttpServer) GetHandlerStaticPage(title string, descr string) http.HandlerFunc {
	handlerName := fmt.Sprintf("GetHandlerStaticPage[%s]", title)
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set(HeaderContentType, MIMEHtml)
		w.WriteHeader(http.StatusOK)
		n, err := fmt.Fprintf(w, getHtmlPage(title, descr))
		if err != nil {
			s.logger.Printf("💥💥 ERROR: [%s]  was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d\n", handlerName, r.URL.Path, r.RemoteAddr, n)
			http.Error(w, "Internal server error. GetHandlerStaticPage was unable to Fprintf", http.StatusInternalServerError)
		}
	}
}
func (s *GoHttpServer) GetTimeHandler() http.HandlerFunc {
	handlerName := "GetTimeHandler"
	s.logger.Printf(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, s.logger)
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
func (s *GoHttpServer) GetWaitHandler(secondsToSleep int) http.HandlerFunc {
	handlerName := "GetWaitHandler"
	s.logger.Printf(initCallMsg, handlerName)
	durationOfSleep := time.Duration(secondsToSleep) * time.Second
	return func(w http.ResponseWriter, r *http.Request) {
		TraceRequest(handlerName, r, s.logger)
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
