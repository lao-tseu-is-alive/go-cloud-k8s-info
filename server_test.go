package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id" param:"id" header:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name" param:"name" header:"name"`
	}
)

const (
	userJSON = `{"id":1,"name":"Carlos GIL"}`
)

func TestErrorConfig_Error(t *testing.T) {
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
				defaultPort: DefaultPort,
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

func TestJSONResponse(t *testing.T) {
	type args struct {
		w      http.ResponseWriter
		r      *http.Request
		result interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			JSONResponse(tt.args.w, tt.args.r, tt.args.result)
		})
	}
}

func Test_waitForShutdown(t *testing.T) {
	type args struct {
		srv *http.Server
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			waitForShutdown(tt.args.srv)
		})
	}
}
