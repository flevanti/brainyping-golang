package internalstatusmonitorapi

import (
	"fmt"
	"net/http"
)

func StartListener(port string, appRole string) {
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{\"status\":\"OK\",\"appRole\":\"%s\"}", appRole)
	})

	go http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
