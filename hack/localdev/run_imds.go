package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/opc/v2/instance/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		data, err := ioutil.ReadFile("hack/localdev/imds_response.json")
		if err != nil {
			log.Printf("Error reading file: %v", err)
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}
		var metadata map[string]interface{}
		if err := json.Unmarshal(data, &metadata); err != nil {
			log.Printf("Error unmarshaling JSON: %v", err)
			http.Error(w, "Invalid JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(metadata); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
		log.Println("Response sent successfully")
	})

	log.Println("Starting mock IMDS server at http://127.0.0.1:8081/opc/v2/instance/")
	if err := http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
