package requests

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"spread-data-storage-practice-1/src/utils/adapters"
	"spread-data-storage-practice-1/src/utils/objects"
	"spread-data-storage-practice-1/src/utils/websocket"
	"strings"
)

type StoreServer struct {
	adapter   *adapters.DatabaseAdapter
	websocket *websocket.ClusterSocket
	manager   *objects.TransactionManager
	journal   *[]objects.Request
}

func CreateStoreServer(adapter *adapters.DatabaseAdapter, socket *websocket.ClusterSocket, manager *objects.TransactionManager, journal *[]objects.Request) *StoreServer {
	return &StoreServer{adapter: adapter, manager: manager, websocket: socket, journal: journal}
}

func (server *StoreServer) GetClock(w http.ResponseWriter, r *http.Request) {
	setSuccessHeader(w)
	makeErrorResponse(w, fmt.Sprintf("%v", server.websocket.GetLogicTime()), http.StatusOK)
}

func (server *StoreServer) RequestValue(w http.ResponseWriter, r *http.Request) {
	setSuccessHeader(w)

	path := r.URL.Path[1:]
	dirs := strings.Split(path, "/")

	if len(dirs) < 2 {
		makeErrorResponse(w, "don't have key", http.StatusBadRequest)
		return
	}

	key := dirs[1]

	if len(key) <= 0 {
		makeErrorResponse(w, "empty key", http.StatusBadRequest)
		return
	}

	fmt.Println(fmt.Sprintf("Request Type: %v Key: %v", r.Method, key))

	switch r.Method {
	case http.MethodPost:
		server.PutValue(w, r, key)
	case http.MethodGet:
		server.GetValue(w, key)
	default:
		makeErrorResponse(w, "method isn't allowed", http.StatusMethodNotAllowed)
	}
}

func (server *StoreServer) GetValue(w http.ResponseWriter, key string) {

	request := objects.Request{Name: "get_value", Args: []string{key}}
	*server.journal = append(*server.journal, request)
	response := server.manager.SendRequest(request)

	if response.Error != nil {
		makeErrorResponse(w, "can't find value with key", http.StatusInternalServerError)
		return
	}

	value := response.Body
	w.Write([]byte(fmt.Sprintf("{\"value\" : \"%v\"}", value)))
}

func (server *StoreServer) PutValue(w http.ResponseWriter, r *http.Request, key string) {

	if !server.websocket.IsLeader() {
		server.redirectHandler(w, r)
		return
	}

	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		makeErrorResponse(w, "can't read body", http.StatusBadRequest)
		return
	}

	value := string(data)

	if len(value) <= 0 {
		makeErrorResponse(w, "empty string", http.StatusBadRequest)
		return
	}

	request := objects.Request{Name: "put_value", Args: []string{key, value}}
	*server.journal = append(*server.journal, request)
	response := server.manager.SendRequest(request)

	if response.Error != nil {
		makeErrorResponse(w, "can't set value", http.StatusInternalServerError)
		return
	}

	makeErrorResponse(w, "success", http.StatusOK)
}

func (server *StoreServer) redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current request URL
	currentURL := r.URL

	// Create a new URL with the same scheme, host, and path but with the new port
	newURL := &url.URL{
		Scheme: "http",
		Host:   "localhost" + ":" + server.websocket.GetLeaderPort(),
		Path:   currentURL.Path,
	}

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a new POST request to the new URL with the same request body
	req, err := http.NewRequest(http.MethodPost, newURL.String(), bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy headers from the original request to the new request
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Perform the POST request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the response status code and headers to the original response
	w.WriteHeader(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy the response body to the original response
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
