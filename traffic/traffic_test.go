package traffic

import (
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFlowControlMiddleware(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(bytes.Repeat([]byte("0"), 1*MB))

	handler := FlowControlMiddleware(500 * KB)
	handlerFunc := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, &buf)
	})

	flowControl := handler(handlerFunc)
	start := time.Now()
	recorder := httptest.NewRecorder()
	flowControl.ServeHTTP(recorder, httptest.NewRequest("GET", "/subdir/asdf", nil))

	delay := time.Now().Sub(start)
	wantDelay := 2 * time.Second
	if math.Abs(wantDelay.Seconds()-delay.Seconds()) > 0.1 {
		t.Errorf("ServeHTTP() delay = %v, wantDelay %v", delay, wantDelay)
		return
	}
	got := recorder.Body.String()
	want := buf.String()
	if !strings.Contains(got, want) {
		t.Fatalf("ServeHTTP() got = %v, want = %v", got, want)
	}
}

func TestFlowControlResponseWriter(t *testing.T) {
	got := FlowControlResponseWriter(nil, context.Background(), 1)
	want := &FlowControlledResponseWriter{
		ResponseWriter: nil,
		ctx:            context.Background(),
		bytesPerSec:    1,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("FlowControlResponseWriter() = %v, want %v", got, want)
	}
}

func TestFlowControlledResponseWriter_Write(t *testing.T) {
	t.Parallel()
	type fields struct {
		ResponseWriter http.ResponseWriter
		ctxProvider    func() context.Context
		bytesPerSec    int
	}
	type args struct {
		p []byte
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      int
		wantDelay time.Duration
		wantErr   bool
	}{
		{
			name: "1MB with 500KB/s",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				ctxProvider:    context.Background,
				bytesPerSec:    500 * KB,
			},
			args:      args{p: bytes.Repeat([]byte("0"), 1*MB)},
			wantDelay: 2 * time.Second,
			want:      1 * MB,
		},
		{
			name: "already closed context",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				ctxProvider: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				},
				bytesPerSec: 500 * KB,
			},
			args:      args{p: bytes.Repeat([]byte("0"), 1*MB)},
			wantErr:   true,
			wantDelay: 0,
			want:      0,
		},
		{
			name: "context cancellation",
			fields: fields{
				ResponseWriter: httptest.NewRecorder(),
				ctxProvider: func() context.Context {
					ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
					return ctx
				},
				bytesPerSec: 500 * KB,
			},
			args:      args{p: bytes.Repeat([]byte("0"), 1*MB)},
			wantErr:   true,
			wantDelay: 500 * time.Millisecond,
			want:      500 * KB,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &FlowControlledResponseWriter{
				ResponseWriter: tt.fields.ResponseWriter,
				ctx:            tt.fields.ctxProvider(),
				bytesPerSec:    tt.fields.bytesPerSec,
			}
			start := time.Now()
			got, err := s.Write(tt.args.p)
			delay := time.Now().Sub(start)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Log(err)
			}
			if math.Abs(tt.wantDelay.Seconds()-delay.Seconds()) > 0.1 {
				t.Errorf("Write() delay = %v, wantDelay %v", delay, tt.wantDelay)
				return
			}
			if got != tt.want {
				t.Errorf("Write() got = %v, want %v", got, tt.want)
			}
		})
	}
}
