package xhttp_test

import (
	"fmt"
	"github.com/nerg4l/goperiment/xhttp"
	"log"
	"net/http"
)

func ExampleUnsafeDynamicHandler() {
	handler := &xhttp.UnsafeDynamicHandler{}
	handler.Append(func(host, path string) http.Handler {
		var id string
		// Sscanf scans space-separated values because of that slash is replaced
		// with space in UnsafeDynamicHandler.
		if _, err := fmt.Sscanf(path, " foo %s", &id); err != nil {
			// If constructed correctly the err will be `input does not match format`
			// when the path does not match the format. When constructed incorrectly
			// it will always fail.
			return nil
		}
		// It is not recommended to do validation of the parameters here and
		// return nil. That wouldn't stop the parameter checks. However,
		// conditions can be added to return different http.Handler.
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Scanned parameters can be accessed here.
			_, _ = fmt.Fprintf(w, "%s", id)
		})
	})
	log.Fatal(http.ListenAndServe(":8080", handler))
}

type H struct {
	id string
}

// ServeDynamic is called on a copy of an empty H. This means all parameters
// has a zero value.
func (h H) ServeDynamic(host, path string) http.Handler {
	// Sscanf scans space-separated values because of that slash is replaced
	// with space in UnsafeDynamicHandler.
	if _, err := fmt.Sscanf(path, " foo %s", &h.id); err != nil {
		// If constructed correctly the err will be `input does not match format`
		// when the path does not match the format. When constructed incorrectly
		// it will always fail.
		return nil
	}
	// It is not recommended to do validation of the parameters here and
	// return nil. That wouldn't stop the parameter checks. However,
	// conditions can be added to return different http.Handler.
	return &h
}

// ServeHTTP can be implemented as value receiver or pointer receiver depending
// on the size of struct H. Variable h in this case will contain a copy of the
// scanned values.
func (h H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "%s", h.id)
}

func ExampleUnsafeDynamicHandler2() {
	handler := &xhttp.UnsafeDynamicHandler{}
	handler.Append(H{}.ServeDynamic)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
