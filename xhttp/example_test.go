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
		// scanf scans space-separated values because of that
		// slash is replaced with space in UnsafeDynamicHandler
		_, err := fmt.Sscanf(path, " foo %s", &id)
		if err != nil {
			return nil
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintf(w, "%s", id)
		})
	})
	log.Fatal(http.ListenAndServe(":8080", handler))
}
