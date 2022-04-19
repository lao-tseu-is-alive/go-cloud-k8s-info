package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

const (
	VERSION                = "0.2.8"
	APP                    = "go-info-server"
	DefaultPort            = 8080
	secondsToSleep         = 10
	secondsShutDownTimeout = 5 // maximum number of second to wait before closing server
)

var logger *log.Logger

type RuntimeInfo struct {
	Hostname     string              `json:"hostname"` //  host name reported by the kernel.
	Pid          int                 `json:"pid"`      //  process id of the caller.
	PPid         int                 `json:"ppid"`     //  process id of the caller's parent.
	Uid          int                 `json:"uid"`      //  numeric user id of the caller.
	Appname      string              `json:"appname"`
	Version      string              `json:"version"`
	ParamName    string              `json:"param_name"`
	RemoteAddr   string              `json:"remote_addr"`
	GOOS         string              `json:"goos"`
	GOARCH       string              `json:"goarch"`
	Runtime      string              `json:"runtime"`
	NumGoroutine string              `json:"num_goroutine"`
	NumCPU       string              `json:"num_cpu"`
	EnvVars      []string            `json:"env_vars"`
	Headers      map[string][]string `json:"headers"`
}

type ErrorConfig struct {
	err error
	msg string
}

func (e *ErrorConfig) Error() string {
	return fmt.Sprintf("%s : %v", e.msg, e.err)
}

//GetPortFromEnv returns a valid TCP/IP listening ':PORT' string based on the values of environment variable :
//	PORT : int value between 1 and 65535 (the parameter defaultPort will be used if env is not defined)
// in case the ENV variable PORT exists and contains an invalid integer the functions returns an empty string and an error
func GetPortFromEnv(defaultPort int) (string, error) {
	srvPort := defaultPort

	var err error
	val, exist := os.LookupEnv("PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			return "", &ErrorConfig{
				err: err,
				msg: "ERROR: CONFIG ENV PORT should contain a valid integer.",
			}
		}
		if srvPort < 1 || srvPort > 65535 {
			return "", &ErrorConfig{
				err: err,
				msg: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
			}
		}
	}
	return fmt.Sprintf(":%d", srvPort), nil
}

func JSONResponse(w http.ResponseWriter, r *http.Request, result interface{}) {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Printf("ERROR: 'JSON marshal failed. Error: %v'", err)
		return
	}
	var prettyOutput bytes.Buffer
	json.Indent(&prettyOutput, body, "", "  ")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(prettyOutput.Bytes())
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan
	logger.Println("INFO: 'Shutting down server...'")

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*secondsShutDownTimeout)
	defer cancel()
	srv.Shutdown(ctx)
	<-ctx.Done()

	logger.Println("INFO: 'Server gracefully stopped'")
	os.Exit(0)
}

func main() {
	listenAddr, err := GetPortFromEnv(DefaultPort)
	if err != nil {
		log.Fatalf("💥💥 error calling GetPortFromEnv. error: %v\n", err)
	}
	//initialize a logger for server messages output
	logger = log.New(os.Stdout, fmt.Sprintf("HTTP_SERVER_%s ", APP), log.LstdFlags)

	logger.Printf("INFO: 'Starting HTTP server on port %s'", listenAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			fmt.Fprintln(w, "User GET")
		}
		if r.Method == http.MethodPost {
			fmt.Fprintln(w, "User POST")
		}
	})

	mux.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			now := time.Now()
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "{\"time\":\"%s\"}", now.Format(time.RFC3339))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	mux.HandleFunc("/wait", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("request: %s '%s'\tremoteAddr: %s,\t sleeping %d seconds\n", r.Method, r.RequestURI, r.RemoteAddr, secondsToSleep)
		if r.Method == http.MethodGet {
			// simulate a delay to be ready
			time.Sleep(secondsToSleep * time.Second)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	mux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("request: %s '%s'\tremoteAddr: %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("request: %s '%s'\tremoteAddr: %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("request: %s '%s'\tremoteAddr: %s\n", r.Method, r.RequestURI, r.RemoteAddr)
		hostName, err := os.Hostname()
		if err != nil {
			logger.Printf("ERROR: 'os.Hostname() returned an error : %v'", err)
			hostName = "#unknown#"
		}
		if r.Method == http.MethodGet {
			query := r.URL.Query()
			name := query.Get("name")
			if name == "" {
				name = "_EMPTY_STRING_"
			}
			data := RuntimeInfo{
				Hostname:     hostName,
				Pid:          os.Getpid(),
				PPid:         os.Getppid(),
				Uid:          os.Getuid(),
				Appname:      APP,
				Version:      VERSION,
				ParamName:    name,
				RemoteAddr:   r.RemoteAddr, // ip address of the original request or the last proxy
				GOOS:         runtime.GOOS,
				GOARCH:       runtime.GOARCH,
				Runtime:      runtime.Version(),
				NumGoroutine: strconv.FormatInt(int64(runtime.NumGoroutine()), 10),
				NumCPU:       strconv.FormatInt(int64(runtime.NumCPU()), 10),
				EnvVars:      os.Environ(),
				Headers:      r.Header,
			}
			JSONResponse(w, r, data)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	// Create server
	srv := &http.Server{
		Handler:      mux,
		Addr:         listenAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		logger.Println("INFO: 'Will start ListenAndServe...'")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("💥💥 ERROR: 'Could not listen on %q: %s'\n", listenAddr, err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}
