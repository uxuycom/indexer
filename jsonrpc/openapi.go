package jsonrpc

import (
	"log"
	"net/http"
	"os"
)

func CreateOpenApi() {
	http.HandleFunc("/v1/docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("docs/openapi.json")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	// start
	if err := http.ListenAndServe(":8011", nil); err != nil {
		log.Fatal(err)
	}
}
