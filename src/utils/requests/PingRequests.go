package requests

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type StatusResponse struct {
	Status string `json:"status"`
}

func setSuccessHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func makeErrorResponse(w http.ResponseWriter, body string, status int) {
	http.Error(w, fmt.Sprintf("{\"status\":\"%v\"}", body), status)
	w.WriteHeader(status)
}
func makeResponse(w http.ResponseWriter, status string) error {
	response := StatusResponse{Status: status}

	jsonBody, err := json.Marshal(response)

	if err != nil {
		makeErrorResponse(w, "Failure", http.StatusBadRequest)
		return errors.New("Can't parse JSON")
	}

	w.Write(jsonBody)
	w.WriteHeader(http.StatusOK)

	return nil
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	http.Error(w, "{\"status\":\"Success\"}", http.StatusOK)
}
